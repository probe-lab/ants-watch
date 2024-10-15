// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// AgentVersion is an object representing the database table.
type AgentVersion struct { // A unique id that identifies a agent version.
	ID int `boil:"id" json:"id" toml:"id" yaml:"id"`
	// Timestamp of when this agent version was seen the last time.
	CreatedAt time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	// Agent version string as reported from the remote peer.
	AgentVersion string `boil:"agent_version" json:"agent_version" toml:"agent_version" yaml:"agent_version"`

	R *agentVersionR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L agentVersionL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var AgentVersionColumns = struct {
	ID           string
	CreatedAt    string
	AgentVersion string
}{
	ID:           "id",
	CreatedAt:    "created_at",
	AgentVersion: "agent_version",
}

var AgentVersionTableColumns = struct {
	ID           string
	CreatedAt    string
	AgentVersion string
}{
	ID:           "agent_versions.id",
	CreatedAt:    "agent_versions.created_at",
	AgentVersion: "agent_versions.agent_version",
}

// Generated where

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var AgentVersionWhere = struct {
	ID           whereHelperint
	CreatedAt    whereHelpertime_Time
	AgentVersion whereHelperstring
}{
	ID:           whereHelperint{field: "\"agent_versions\".\"id\""},
	CreatedAt:    whereHelpertime_Time{field: "\"agent_versions\".\"created_at\""},
	AgentVersion: whereHelperstring{field: "\"agent_versions\".\"agent_version\""},
}

// AgentVersionRels is where relationship names are stored.
var AgentVersionRels = struct {
	Peers string
}{
	Peers: "Peers",
}

// agentVersionR is where relationships are stored.
type agentVersionR struct {
	Peers PeerSlice `boil:"Peers" json:"Peers" toml:"Peers" yaml:"Peers"`
}

// NewStruct creates a new relationship struct
func (*agentVersionR) NewStruct() *agentVersionR {
	return &agentVersionR{}
}

func (r *agentVersionR) GetPeers() PeerSlice {
	if r == nil {
		return nil
	}
	return r.Peers
}

// agentVersionL is where Load methods for each relationship are stored.
type agentVersionL struct{}

var (
	agentVersionAllColumns            = []string{"id", "created_at", "agent_version"}
	agentVersionColumnsWithoutDefault = []string{"created_at", "agent_version"}
	agentVersionColumnsWithDefault    = []string{"id"}
	agentVersionPrimaryKeyColumns     = []string{"id"}
	agentVersionGeneratedColumns      = []string{"id"}
)

type (
	// AgentVersionSlice is an alias for a slice of pointers to AgentVersion.
	// This should almost always be used instead of []AgentVersion.
	AgentVersionSlice []*AgentVersion
	// AgentVersionHook is the signature for custom AgentVersion hook methods
	AgentVersionHook func(context.Context, boil.ContextExecutor, *AgentVersion) error

	agentVersionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	agentVersionType                 = reflect.TypeOf(&AgentVersion{})
	agentVersionMapping              = queries.MakeStructMapping(agentVersionType)
	agentVersionPrimaryKeyMapping, _ = queries.BindMapping(agentVersionType, agentVersionMapping, agentVersionPrimaryKeyColumns)
	agentVersionInsertCacheMut       sync.RWMutex
	agentVersionInsertCache          = make(map[string]insertCache)
	agentVersionUpdateCacheMut       sync.RWMutex
	agentVersionUpdateCache          = make(map[string]updateCache)
	agentVersionUpsertCacheMut       sync.RWMutex
	agentVersionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var agentVersionAfterSelectHooks []AgentVersionHook

var agentVersionBeforeInsertHooks []AgentVersionHook
var agentVersionAfterInsertHooks []AgentVersionHook

var agentVersionBeforeUpdateHooks []AgentVersionHook
var agentVersionAfterUpdateHooks []AgentVersionHook

var agentVersionBeforeDeleteHooks []AgentVersionHook
var agentVersionAfterDeleteHooks []AgentVersionHook

var agentVersionBeforeUpsertHooks []AgentVersionHook
var agentVersionAfterUpsertHooks []AgentVersionHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *AgentVersion) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *AgentVersion) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *AgentVersion) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *AgentVersion) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *AgentVersion) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *AgentVersion) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *AgentVersion) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *AgentVersion) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *AgentVersion) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range agentVersionAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddAgentVersionHook registers your hook function for all future operations.
func AddAgentVersionHook(hookPoint boil.HookPoint, agentVersionHook AgentVersionHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		agentVersionAfterSelectHooks = append(agentVersionAfterSelectHooks, agentVersionHook)
	case boil.BeforeInsertHook:
		agentVersionBeforeInsertHooks = append(agentVersionBeforeInsertHooks, agentVersionHook)
	case boil.AfterInsertHook:
		agentVersionAfterInsertHooks = append(agentVersionAfterInsertHooks, agentVersionHook)
	case boil.BeforeUpdateHook:
		agentVersionBeforeUpdateHooks = append(agentVersionBeforeUpdateHooks, agentVersionHook)
	case boil.AfterUpdateHook:
		agentVersionAfterUpdateHooks = append(agentVersionAfterUpdateHooks, agentVersionHook)
	case boil.BeforeDeleteHook:
		agentVersionBeforeDeleteHooks = append(agentVersionBeforeDeleteHooks, agentVersionHook)
	case boil.AfterDeleteHook:
		agentVersionAfterDeleteHooks = append(agentVersionAfterDeleteHooks, agentVersionHook)
	case boil.BeforeUpsertHook:
		agentVersionBeforeUpsertHooks = append(agentVersionBeforeUpsertHooks, agentVersionHook)
	case boil.AfterUpsertHook:
		agentVersionAfterUpsertHooks = append(agentVersionAfterUpsertHooks, agentVersionHook)
	}
}

// One returns a single agentVersion record from the query.
func (q agentVersionQuery) One(ctx context.Context, exec boil.ContextExecutor) (*AgentVersion, error) {
	o := &AgentVersion{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for agent_versions")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all AgentVersion records from the query.
func (q agentVersionQuery) All(ctx context.Context, exec boil.ContextExecutor) (AgentVersionSlice, error) {
	var o []*AgentVersion

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to AgentVersion slice")
	}

	if len(agentVersionAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all AgentVersion records in the query.
func (q agentVersionQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count agent_versions rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q agentVersionQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if agent_versions exists")
	}

	return count > 0, nil
}

// Peers retrieves all the peer's Peers with an executor.
func (o *AgentVersion) Peers(mods ...qm.QueryMod) peerQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"peers\".\"agent_version_id\"=?", o.ID),
	)

	return Peers(queryMods...)
}

// LoadPeers allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (agentVersionL) LoadPeers(ctx context.Context, e boil.ContextExecutor, singular bool, maybeAgentVersion interface{}, mods queries.Applicator) error {
	var slice []*AgentVersion
	var object *AgentVersion

	if singular {
		var ok bool
		object, ok = maybeAgentVersion.(*AgentVersion)
		if !ok {
			object = new(AgentVersion)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeAgentVersion)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeAgentVersion))
			}
		}
	} else {
		s, ok := maybeAgentVersion.(*[]*AgentVersion)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeAgentVersion)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeAgentVersion))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &agentVersionR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &agentVersionR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.ID) {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`peers`),
		qm.WhereIn(`peers.agent_version_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load peers")
	}

	var resultSlice []*Peer
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice peers")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on peers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for peers")
	}

	if len(peerAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Peers = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &peerR{}
			}
			foreign.R.AgentVersion = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if queries.Equal(local.ID, foreign.AgentVersionID) {
				local.R.Peers = append(local.R.Peers, foreign)
				if foreign.R == nil {
					foreign.R = &peerR{}
				}
				foreign.R.AgentVersion = local
				break
			}
		}
	}

	return nil
}

// AddPeers adds the given related objects to the existing relationships
// of the agent_version, optionally inserting them as new records.
// Appends related to o.R.Peers.
// Sets related.R.AgentVersion appropriately.
func (o *AgentVersion) AddPeers(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Peer) error {
	var err error
	for _, rel := range related {
		if insert {
			queries.Assign(&rel.AgentVersionID, o.ID)
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"peers\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"agent_version_id"}),
				strmangle.WhereClause("\"", "\"", 2, peerPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			queries.Assign(&rel.AgentVersionID, o.ID)
		}
	}

	if o.R == nil {
		o.R = &agentVersionR{
			Peers: related,
		}
	} else {
		o.R.Peers = append(o.R.Peers, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &peerR{
				AgentVersion: o,
			}
		} else {
			rel.R.AgentVersion = o
		}
	}
	return nil
}

// SetPeers removes all previously related items of the
// agent_version replacing them completely with the passed
// in related items, optionally inserting them as new records.
// Sets o.R.AgentVersion's Peers accordingly.
// Replaces o.R.Peers with related.
// Sets related.R.AgentVersion's Peers accordingly.
func (o *AgentVersion) SetPeers(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Peer) error {
	query := "update \"peers\" set \"agent_version_id\" = null where \"agent_version_id\" = $1"
	values := []interface{}{o.ID}
	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, query)
		fmt.Fprintln(writer, values)
	}
	_, err := exec.ExecContext(ctx, query, values...)
	if err != nil {
		return errors.Wrap(err, "failed to remove relationships before set")
	}

	if o.R != nil {
		for _, rel := range o.R.Peers {
			queries.SetScanner(&rel.AgentVersionID, nil)
			if rel.R == nil {
				continue
			}

			rel.R.AgentVersion = nil
		}
		o.R.Peers = nil
	}

	return o.AddPeers(ctx, exec, insert, related...)
}

// RemovePeers relationships from objects passed in.
// Removes related items from R.Peers (uses pointer comparison, removal does not keep order)
// Sets related.R.AgentVersion.
func (o *AgentVersion) RemovePeers(ctx context.Context, exec boil.ContextExecutor, related ...*Peer) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	for _, rel := range related {
		queries.SetScanner(&rel.AgentVersionID, nil)
		if rel.R != nil {
			rel.R.AgentVersion = nil
		}
		if _, err = rel.Update(ctx, exec, boil.Whitelist("agent_version_id")); err != nil {
			return err
		}
	}
	if o.R == nil {
		return nil
	}

	for _, rel := range related {
		for i, ri := range o.R.Peers {
			if rel != ri {
				continue
			}

			ln := len(o.R.Peers)
			if ln > 1 && i < ln-1 {
				o.R.Peers[i] = o.R.Peers[ln-1]
			}
			o.R.Peers = o.R.Peers[:ln-1]
			break
		}
	}

	return nil
}

// AgentVersions retrieves all the records using an executor.
func AgentVersions(mods ...qm.QueryMod) agentVersionQuery {
	mods = append(mods, qm.From("\"agent_versions\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"agent_versions\".*"})
	}

	return agentVersionQuery{q}
}

// FindAgentVersion retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindAgentVersion(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*AgentVersion, error) {
	agentVersionObj := &AgentVersion{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"agent_versions\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, agentVersionObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from agent_versions")
	}

	if err = agentVersionObj.doAfterSelectHooks(ctx, exec); err != nil {
		return agentVersionObj, err
	}

	return agentVersionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *AgentVersion) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no agent_versions provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(agentVersionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	agentVersionInsertCacheMut.RLock()
	cache, cached := agentVersionInsertCache[key]
	agentVersionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			agentVersionAllColumns,
			agentVersionColumnsWithDefault,
			agentVersionColumnsWithoutDefault,
			nzDefaults,
		)
		wl = strmangle.SetComplement(wl, agentVersionGeneratedColumns)

		cache.valueMapping, err = queries.BindMapping(agentVersionType, agentVersionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(agentVersionType, agentVersionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"agent_versions\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"agent_versions\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into agent_versions")
	}

	if !cached {
		agentVersionInsertCacheMut.Lock()
		agentVersionInsertCache[key] = cache
		agentVersionInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the AgentVersion.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *AgentVersion) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	agentVersionUpdateCacheMut.RLock()
	cache, cached := agentVersionUpdateCache[key]
	agentVersionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			agentVersionAllColumns,
			agentVersionPrimaryKeyColumns,
		)
		wl = strmangle.SetComplement(wl, agentVersionGeneratedColumns)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update agent_versions, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"agent_versions\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, agentVersionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(agentVersionType, agentVersionMapping, append(wl, agentVersionPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update agent_versions row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for agent_versions")
	}

	if !cached {
		agentVersionUpdateCacheMut.Lock()
		agentVersionUpdateCache[key] = cache
		agentVersionUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q agentVersionQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for agent_versions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for agent_versions")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o AgentVersionSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), agentVersionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"agent_versions\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, agentVersionPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in agentVersion slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all agentVersion")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *AgentVersion) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no agent_versions provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(agentVersionColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	agentVersionUpsertCacheMut.RLock()
	cache, cached := agentVersionUpsertCache[key]
	agentVersionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			agentVersionAllColumns,
			agentVersionColumnsWithDefault,
			agentVersionColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			agentVersionAllColumns,
			agentVersionPrimaryKeyColumns,
		)

		insert = strmangle.SetComplement(insert, agentVersionGeneratedColumns)
		update = strmangle.SetComplement(update, agentVersionGeneratedColumns)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert agent_versions, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(agentVersionPrimaryKeyColumns))
			copy(conflict, agentVersionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"agent_versions\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(agentVersionType, agentVersionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(agentVersionType, agentVersionMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert agent_versions")
	}

	if !cached {
		agentVersionUpsertCacheMut.Lock()
		agentVersionUpsertCache[key] = cache
		agentVersionUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single AgentVersion record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *AgentVersion) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no AgentVersion provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), agentVersionPrimaryKeyMapping)
	sql := "DELETE FROM \"agent_versions\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from agent_versions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for agent_versions")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q agentVersionQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no agentVersionQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from agent_versions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for agent_versions")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o AgentVersionSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(agentVersionBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), agentVersionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"agent_versions\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, agentVersionPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from agentVersion slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for agent_versions")
	}

	if len(agentVersionAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *AgentVersion) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindAgentVersion(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *AgentVersionSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := AgentVersionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), agentVersionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"agent_versions\".* FROM \"agent_versions\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, agentVersionPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in AgentVersionSlice")
	}

	*o = slice

	return nil
}

// AgentVersionExists checks if the AgentVersion row exists.
func AgentVersionExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"agent_versions\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if agent_versions exists")
	}

	return exists, nil
}
