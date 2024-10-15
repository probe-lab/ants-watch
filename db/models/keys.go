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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Key is an object representing the database table.
type Key struct {
	ID        int         `boil:"id" json:"id" toml:"id" yaml:"id"`
	PeerID    null.Int64  `boil:"peer_id" json:"peer_id,omitempty" toml:"peer_id" yaml:"peer_id,omitempty"`
	MultiHash null.String `boil:"multi_hash" json:"multi_hash,omitempty" toml:"multi_hash" yaml:"multi_hash,omitempty"`

	R *keyR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L keyL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var KeyColumns = struct {
	ID        string
	PeerID    string
	MultiHash string
}{
	ID:        "id",
	PeerID:    "peer_id",
	MultiHash: "multi_hash",
}

var KeyTableColumns = struct {
	ID        string
	PeerID    string
	MultiHash string
}{
	ID:        "keys.id",
	PeerID:    "keys.peer_id",
	MultiHash: "keys.multi_hash",
}

// Generated where

type whereHelpernull_Int64 struct{ field string }

func (w whereHelpernull_Int64) EQ(x null.Int64) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Int64) NEQ(x null.Int64) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Int64) LT(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Int64) LTE(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Int64) GT(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Int64) GTE(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}
func (w whereHelpernull_Int64) IN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelpernull_Int64) NIN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

func (w whereHelpernull_Int64) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Int64) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var KeyWhere = struct {
	ID        whereHelperint
	PeerID    whereHelpernull_Int64
	MultiHash whereHelpernull_String
}{
	ID:        whereHelperint{field: "\"keys\".\"id\""},
	PeerID:    whereHelpernull_Int64{field: "\"keys\".\"peer_id\""},
	MultiHash: whereHelpernull_String{field: "\"keys\".\"multi_hash\""},
}

// KeyRels is where relationship names are stored.
var KeyRels = struct {
	Peer string
}{
	Peer: "Peer",
}

// keyR is where relationships are stored.
type keyR struct {
	Peer *Peer `boil:"Peer" json:"Peer" toml:"Peer" yaml:"Peer"`
}

// NewStruct creates a new relationship struct
func (*keyR) NewStruct() *keyR {
	return &keyR{}
}

func (r *keyR) GetPeer() *Peer {
	if r == nil {
		return nil
	}
	return r.Peer
}

// keyL is where Load methods for each relationship are stored.
type keyL struct{}

var (
	keyAllColumns            = []string{"id", "peer_id", "multi_hash"}
	keyColumnsWithoutDefault = []string{}
	keyColumnsWithDefault    = []string{"id", "peer_id", "multi_hash"}
	keyPrimaryKeyColumns     = []string{"id"}
	keyGeneratedColumns      = []string{"id"}
)

type (
	// KeySlice is an alias for a slice of pointers to Key.
	// This should almost always be used instead of []Key.
	KeySlice []*Key
	// KeyHook is the signature for custom Key hook methods
	KeyHook func(context.Context, boil.ContextExecutor, *Key) error

	keyQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	keyType                 = reflect.TypeOf(&Key{})
	keyMapping              = queries.MakeStructMapping(keyType)
	keyPrimaryKeyMapping, _ = queries.BindMapping(keyType, keyMapping, keyPrimaryKeyColumns)
	keyInsertCacheMut       sync.RWMutex
	keyInsertCache          = make(map[string]insertCache)
	keyUpdateCacheMut       sync.RWMutex
	keyUpdateCache          = make(map[string]updateCache)
	keyUpsertCacheMut       sync.RWMutex
	keyUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var keyAfterSelectHooks []KeyHook

var keyBeforeInsertHooks []KeyHook
var keyAfterInsertHooks []KeyHook

var keyBeforeUpdateHooks []KeyHook
var keyAfterUpdateHooks []KeyHook

var keyBeforeDeleteHooks []KeyHook
var keyAfterDeleteHooks []KeyHook

var keyBeforeUpsertHooks []KeyHook
var keyAfterUpsertHooks []KeyHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Key) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Key) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Key) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Key) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Key) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Key) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Key) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Key) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Key) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range keyAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddKeyHook registers your hook function for all future operations.
func AddKeyHook(hookPoint boil.HookPoint, keyHook KeyHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		keyAfterSelectHooks = append(keyAfterSelectHooks, keyHook)
	case boil.BeforeInsertHook:
		keyBeforeInsertHooks = append(keyBeforeInsertHooks, keyHook)
	case boil.AfterInsertHook:
		keyAfterInsertHooks = append(keyAfterInsertHooks, keyHook)
	case boil.BeforeUpdateHook:
		keyBeforeUpdateHooks = append(keyBeforeUpdateHooks, keyHook)
	case boil.AfterUpdateHook:
		keyAfterUpdateHooks = append(keyAfterUpdateHooks, keyHook)
	case boil.BeforeDeleteHook:
		keyBeforeDeleteHooks = append(keyBeforeDeleteHooks, keyHook)
	case boil.AfterDeleteHook:
		keyAfterDeleteHooks = append(keyAfterDeleteHooks, keyHook)
	case boil.BeforeUpsertHook:
		keyBeforeUpsertHooks = append(keyBeforeUpsertHooks, keyHook)
	case boil.AfterUpsertHook:
		keyAfterUpsertHooks = append(keyAfterUpsertHooks, keyHook)
	}
}

// One returns a single key record from the query.
func (q keyQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Key, error) {
	o := &Key{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for keys")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Key records from the query.
func (q keyQuery) All(ctx context.Context, exec boil.ContextExecutor) (KeySlice, error) {
	var o []*Key

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Key slice")
	}

	if len(keyAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Key records in the query.
func (q keyQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count keys rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q keyQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if keys exists")
	}

	return count > 0, nil
}

// Peer pointed to by the foreign key.
func (o *Key) Peer(mods ...qm.QueryMod) peerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PeerID),
	}

	queryMods = append(queryMods, mods...)

	return Peers(queryMods...)
}

// LoadPeer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (keyL) LoadPeer(ctx context.Context, e boil.ContextExecutor, singular bool, maybeKey interface{}, mods queries.Applicator) error {
	var slice []*Key
	var object *Key

	if singular {
		var ok bool
		object, ok = maybeKey.(*Key)
		if !ok {
			object = new(Key)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeKey)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeKey))
			}
		}
	} else {
		s, ok := maybeKey.(*[]*Key)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeKey)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeKey))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &keyR{}
		}
		if !queries.IsNil(object.PeerID) {
			args = append(args, object.PeerID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &keyR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.PeerID) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.PeerID) {
				args = append(args, obj.PeerID)
			}

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`peers`),
		qm.WhereIn(`peers.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Peer")
	}

	var resultSlice []*Peer
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Peer")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for peers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for peers")
	}

	if len(keyAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Peer = foreign
		if foreign.R == nil {
			foreign.R = &peerR{}
		}
		foreign.R.Keys = append(foreign.R.Keys, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.PeerID, foreign.ID) {
				local.R.Peer = foreign
				if foreign.R == nil {
					foreign.R = &peerR{}
				}
				foreign.R.Keys = append(foreign.R.Keys, local)
				break
			}
		}
	}

	return nil
}

// SetPeer of the key to the related item.
// Sets o.R.Peer to related.
// Adds o to related.R.Keys.
func (o *Key) SetPeer(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Peer) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"keys\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"peer_id"}),
		strmangle.WhereClause("\"", "\"", 2, keyPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.PeerID, related.ID)
	if o.R == nil {
		o.R = &keyR{
			Peer: related,
		}
	} else {
		o.R.Peer = related
	}

	if related.R == nil {
		related.R = &peerR{
			Keys: KeySlice{o},
		}
	} else {
		related.R.Keys = append(related.R.Keys, o)
	}

	return nil
}

// RemovePeer relationship.
// Sets o.R.Peer to nil.
// Removes o from all passed in related items' relationships struct.
func (o *Key) RemovePeer(ctx context.Context, exec boil.ContextExecutor, related *Peer) error {
	var err error

	queries.SetScanner(&o.PeerID, nil)
	if _, err = o.Update(ctx, exec, boil.Whitelist("peer_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.Peer = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.Keys {
		if queries.Equal(o.PeerID, ri.PeerID) {
			continue
		}

		ln := len(related.R.Keys)
		if ln > 1 && i < ln-1 {
			related.R.Keys[i] = related.R.Keys[ln-1]
		}
		related.R.Keys = related.R.Keys[:ln-1]
		break
	}
	return nil
}

// Keys retrieves all the records using an executor.
func Keys(mods ...qm.QueryMod) keyQuery {
	mods = append(mods, qm.From("\"keys\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"keys\".*"})
	}

	return keyQuery{q}
}

// FindKey retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindKey(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Key, error) {
	keyObj := &Key{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"keys\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, keyObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from keys")
	}

	if err = keyObj.doAfterSelectHooks(ctx, exec); err != nil {
		return keyObj, err
	}

	return keyObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Key) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no keys provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(keyColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	keyInsertCacheMut.RLock()
	cache, cached := keyInsertCache[key]
	keyInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			keyAllColumns,
			keyColumnsWithDefault,
			keyColumnsWithoutDefault,
			nzDefaults,
		)
		wl = strmangle.SetComplement(wl, keyGeneratedColumns)

		cache.valueMapping, err = queries.BindMapping(keyType, keyMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(keyType, keyMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"keys\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"keys\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into keys")
	}

	if !cached {
		keyInsertCacheMut.Lock()
		keyInsertCache[key] = cache
		keyInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Key.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Key) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	keyUpdateCacheMut.RLock()
	cache, cached := keyUpdateCache[key]
	keyUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			keyAllColumns,
			keyPrimaryKeyColumns,
		)
		wl = strmangle.SetComplement(wl, keyGeneratedColumns)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update keys, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"keys\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, keyPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(keyType, keyMapping, append(wl, keyPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update keys row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for keys")
	}

	if !cached {
		keyUpdateCacheMut.Lock()
		keyUpdateCache[key] = cache
		keyUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q keyQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for keys")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for keys")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o KeySlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), keyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"keys\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, keyPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in key slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all key")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Key) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no keys provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(keyColumnsWithDefault, o)

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

	keyUpsertCacheMut.RLock()
	cache, cached := keyUpsertCache[key]
	keyUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			keyAllColumns,
			keyColumnsWithDefault,
			keyColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			keyAllColumns,
			keyPrimaryKeyColumns,
		)

		insert = strmangle.SetComplement(insert, keyGeneratedColumns)
		update = strmangle.SetComplement(update, keyGeneratedColumns)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert keys, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(keyPrimaryKeyColumns))
			copy(conflict, keyPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"keys\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(keyType, keyMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(keyType, keyMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert keys")
	}

	if !cached {
		keyUpsertCacheMut.Lock()
		keyUpsertCache[key] = cache
		keyUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Key record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Key) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Key provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), keyPrimaryKeyMapping)
	sql := "DELETE FROM \"keys\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from keys")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for keys")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q keyQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no keyQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from keys")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for keys")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o KeySlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(keyBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), keyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"keys\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, keyPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from key slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for keys")
	}

	if len(keyAfterDeleteHooks) != 0 {
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
func (o *Key) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindKey(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *KeySlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := KeySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), keyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"keys\".* FROM \"keys\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, keyPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in KeySlice")
	}

	*o = slice

	return nil
}

// KeyExists checks if the Key row exists.
func KeyExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"keys\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if keys exists")
	}

	return exists, nil
}
