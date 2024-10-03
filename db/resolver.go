package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/maxmind"
	"github.com/dennis-tra/nebula-crawler/udger"
	"github.com/probe-lab/ants-watch/db/models"
)

func Resolve(ctx context.Context, dbh *sql.DB, mmc *maxmind.Client, uclient *udger.Client, dbmaddrs models.MultiAddressSlice) error {
	log.WithField("size", len(dbmaddrs)).Infoln("Resolving batch of multi addresses...")

	for _, dbmaddr := range dbmaddrs {
		if err := resolveAddr(ctx, dbh, mmc, uclient, dbmaddr); err != nil {
			log.WithField("maddr", dbmaddr.Maddr).WithError(err).Warnln("Error resolving multi address")
		}
	}

	return nil
}

func resolveAddr(ctx context.Context, dbh *sql.DB, mmc *maxmind.Client, uclient *udger.Client, dbmaddr *models.MultiAddress) error {
	logEntry := log.WithField("maddr", dbmaddr.Maddr)
	txn, err := dbh.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin txn: %w", err)
	}
	defer db.Rollback(txn)

	maddr, err := ma.NewMultiaddr(dbmaddr.Maddr)
	if err != nil {
		logEntry.WithError(err).Warnln("Error parsing multi address - deleting row")
		if _, delErr := dbmaddr.Delete(ctx, txn); err != nil {
			logEntry.WithError(delErr).Warnln("Error deleting multi address")
			return fmt.Errorf("parse multi address: %w", err)
		} else {
			return txn.Commit()
		}
	}

	dbmaddr.Resolved = true
	dbmaddr.IsPublic = null.BoolFrom(manet.IsPublicAddr(maddr))
	dbmaddr.IsRelay = null.BoolFrom(isRelayedMaddr(maddr))

	addrInfos, err := mmc.MaddrInfo(ctx, maddr)
	if err != nil {
		logEntry.WithError(err).Warnln("Error deriving address information from maddr ", maddr)
	}

	if len(addrInfos) == 0 {
		dbmaddr.HasManyAddrs = null.BoolFrom(false)
	} else if len(addrInfos) == 1 {
		dbmaddr.HasManyAddrs = null.BoolFrom(false)

		// we only have one addrInfo, extract it from the map
		var addr string
		var addrInfo *maxmind.AddrInfo
		for k, v := range addrInfos {
			addr, addrInfo = k, v
			break
		}

		dbmaddr.Asn = null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0)
		dbmaddr.Country = null.NewString(addrInfo.Country, addrInfo.Country != "")
		dbmaddr.Continent = null.NewString(addrInfo.Continent, addrInfo.Continent != "")
		dbmaddr.Addr = null.NewString(addr, addr != "")

		if uclient != nil {
			datacenterID, err := uclient.Datacenter(addr)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logEntry.WithError(err).WithField("addr", addr).Warnln("Error resolving ip address to datacenter")
			}
			dbmaddr.IsCloud = null.NewInt(datacenterID, datacenterID != 0)
		}

	} else if len(addrInfos) > 1 { // not "else" because the MaddrInfo could have failed and we still want to update the maddr
		dbmaddr.HasManyAddrs = null.BoolFrom(true)

		// Due to dnsaddr protocols each multi address can point to multiple
		// IP addresses each in a different country.
		for addr, addrInfo := range addrInfos {
			datacenterID := 0
			if uclient != nil {
				datacenterID, err = uclient.Datacenter(addr)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logEntry.WithError(err).WithField("addr", addr).Warnln("Error resolving ip address to datacenter")
				} else if datacenterID > 0 {
					dbmaddr.IsCloud = null.IntFrom(datacenterID)
				}
			}

			// Save the IP address + country information + asn information
			ipaddr := &models.IPAddress{
				Asn:       null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0),
				IsCloud:   null.NewInt(datacenterID, datacenterID != 0),
				Country:   null.NewString(addrInfo.Country, addrInfo.Country != ""),
				Continent: null.NewString(addrInfo.Continent, addrInfo.Continent != ""),
				Address:   addr,
			}
			if err := dbmaddr.AddIPAddresses(ctx, txn, true, ipaddr); err != nil {
				logEntry.WithError(err).WithField("addr", ipaddr.Address).Warnln("Could not insert ip address")
				return fmt.Errorf("add ip addresses: %w", err)
			}
		}
	}

	if _, err = dbmaddr.Update(ctx, txn, boil.Infer()); err != nil {
		logEntry.WithError(err).Warnln("Could not update multi address")
		return fmt.Errorf("update multi address: %w", err)
	}

	return txn.Commit()
}

func isRelayedMaddr(maddr ma.Multiaddr) bool {
	_, err := maddr.ValueForProtocol(ma.P_CIRCUIT)
	if err == nil {
		return true
	} else if errors.Is(err, ma.ErrProtocolNotFound) {
		return false
	} else {
		log.WithError(err).WithField("maddr", maddr).Warnln("Unexpected error while parsing multi address")
		return false
	}
}
