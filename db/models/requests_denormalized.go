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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// RequestsDenormalized is an object representing the database table.
type RequestsDenormalized struct {
	ID              int               `boil:"id" json:"id" toml:"id" yaml:"id"`
	Timestamp       time.Time         `boil:"timestamp" json:"timestamp" toml:"timestamp" yaml:"timestamp"`
	RequestType     string            `boil:"request_type" json:"request_type" toml:"request_type" yaml:"request_type"`
	AntID           string            `boil:"ant_id" json:"ant_id" toml:"ant_id" yaml:"ant_id"`
	PeerID          string            `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	KeyID           string            `boil:"key_id" json:"key_id" toml:"key_id" yaml:"key_id"`
	MultiAddressIds types.StringArray `boil:"multi_address_ids" json:"multi_address_ids,omitempty" toml:"multi_address_ids" yaml:"multi_address_ids,omitempty"`

	R *requestsDenormalizedR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L requestsDenormalizedL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var RequestsDenormalizedColumns = struct {
	ID              string
	Timestamp       string
	RequestType     string
	AntID           string
	PeerID          string
	KeyID           string
	MultiAddressIds string
}{
	ID:              "id",
	Timestamp:       "timestamp",
	RequestType:     "request_type",
	AntID:           "ant_id",
	PeerID:          "peer_id",
	KeyID:           "key_id",
	MultiAddressIds: "multi_address_ids",
}

var RequestsDenormalizedTableColumns = struct {
	ID              string
	Timestamp       string
	RequestType     string
	AntID           string
	PeerID          string
	KeyID           string
	MultiAddressIds string
}{
	ID:              "requests_denormalized.id",
	Timestamp:       "requests_denormalized.timestamp",
	RequestType:     "requests_denormalized.request_type",
	AntID:           "requests_denormalized.ant_id",
	PeerID:          "requests_denormalized.peer_id",
	KeyID:           "requests_denormalized.key_id",
	MultiAddressIds: "requests_denormalized.multi_address_ids",
}

// Generated where

type whereHelpertypes_StringArray struct{ field string }

func (w whereHelpertypes_StringArray) EQ(x types.StringArray) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpertypes_StringArray) NEQ(x types.StringArray) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpertypes_StringArray) LT(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_StringArray) LTE(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_StringArray) GT(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_StringArray) GTE(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpertypes_StringArray) IsNull() qm.QueryMod { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpertypes_StringArray) IsNotNull() qm.QueryMod {
	return qmhelper.WhereIsNotNull(w.field)
}

var RequestsDenormalizedWhere = struct {
	ID              whereHelperint
	Timestamp       whereHelpertime_Time
	RequestType     whereHelperstring
	AntID           whereHelperstring
	PeerID          whereHelperstring
	KeyID           whereHelperstring
	MultiAddressIds whereHelpertypes_StringArray
}{
	ID:              whereHelperint{field: "\"requests_denormalized\".\"id\""},
	Timestamp:       whereHelpertime_Time{field: "\"requests_denormalized\".\"timestamp\""},
	RequestType:     whereHelperstring{field: "\"requests_denormalized\".\"request_type\""},
	AntID:           whereHelperstring{field: "\"requests_denormalized\".\"ant_id\""},
	PeerID:          whereHelperstring{field: "\"requests_denormalized\".\"peer_id\""},
	KeyID:           whereHelperstring{field: "\"requests_denormalized\".\"key_id\""},
	MultiAddressIds: whereHelpertypes_StringArray{field: "\"requests_denormalized\".\"multi_address_ids\""},
}

// RequestsDenormalizedRels is where relationship names are stored.
var RequestsDenormalizedRels = struct {
}{}

// requestsDenormalizedR is where relationships are stored.
type requestsDenormalizedR struct {
}

// NewStruct creates a new relationship struct
func (*requestsDenormalizedR) NewStruct() *requestsDenormalizedR {
	return &requestsDenormalizedR{}
}

// requestsDenormalizedL is where Load methods for each relationship are stored.
type requestsDenormalizedL struct{}

var (
	requestsDenormalizedAllColumns            = []string{"id", "timestamp", "request_type", "ant_id", "peer_id", "key_id", "multi_address_ids"}
	requestsDenormalizedColumnsWithoutDefault = []string{"timestamp", "request_type", "ant_id", "peer_id", "key_id"}
	requestsDenormalizedColumnsWithDefault    = []string{"id", "multi_address_ids"}
	requestsDenormalizedPrimaryKeyColumns     = []string{"id", "timestamp"}
	requestsDenormalizedGeneratedColumns      = []string{"id"}
)

type (
	// RequestsDenormalizedSlice is an alias for a slice of pointers to RequestsDenormalized.
	// This should almost always be used instead of []RequestsDenormalized.
	RequestsDenormalizedSlice []*RequestsDenormalized
	// RequestsDenormalizedHook is the signature for custom RequestsDenormalized hook methods
	RequestsDenormalizedHook func(context.Context, boil.ContextExecutor, *RequestsDenormalized) error

	requestsDenormalizedQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	requestsDenormalizedType                 = reflect.TypeOf(&RequestsDenormalized{})
	requestsDenormalizedMapping              = queries.MakeStructMapping(requestsDenormalizedType)
	requestsDenormalizedPrimaryKeyMapping, _ = queries.BindMapping(requestsDenormalizedType, requestsDenormalizedMapping, requestsDenormalizedPrimaryKeyColumns)
	requestsDenormalizedInsertCacheMut       sync.RWMutex
	requestsDenormalizedInsertCache          = make(map[string]insertCache)
	requestsDenormalizedUpdateCacheMut       sync.RWMutex
	requestsDenormalizedUpdateCache          = make(map[string]updateCache)
	requestsDenormalizedUpsertCacheMut       sync.RWMutex
	requestsDenormalizedUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var requestsDenormalizedAfterSelectHooks []RequestsDenormalizedHook

var requestsDenormalizedBeforeInsertHooks []RequestsDenormalizedHook
var requestsDenormalizedAfterInsertHooks []RequestsDenormalizedHook

var requestsDenormalizedBeforeUpdateHooks []RequestsDenormalizedHook
var requestsDenormalizedAfterUpdateHooks []RequestsDenormalizedHook

var requestsDenormalizedBeforeDeleteHooks []RequestsDenormalizedHook
var requestsDenormalizedAfterDeleteHooks []RequestsDenormalizedHook

var requestsDenormalizedBeforeUpsertHooks []RequestsDenormalizedHook
var requestsDenormalizedAfterUpsertHooks []RequestsDenormalizedHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *RequestsDenormalized) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *RequestsDenormalized) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *RequestsDenormalized) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *RequestsDenormalized) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *RequestsDenormalized) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *RequestsDenormalized) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *RequestsDenormalized) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *RequestsDenormalized) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *RequestsDenormalized) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range requestsDenormalizedAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddRequestsDenormalizedHook registers your hook function for all future operations.
func AddRequestsDenormalizedHook(hookPoint boil.HookPoint, requestsDenormalizedHook RequestsDenormalizedHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		requestsDenormalizedAfterSelectHooks = append(requestsDenormalizedAfterSelectHooks, requestsDenormalizedHook)
	case boil.BeforeInsertHook:
		requestsDenormalizedBeforeInsertHooks = append(requestsDenormalizedBeforeInsertHooks, requestsDenormalizedHook)
	case boil.AfterInsertHook:
		requestsDenormalizedAfterInsertHooks = append(requestsDenormalizedAfterInsertHooks, requestsDenormalizedHook)
	case boil.BeforeUpdateHook:
		requestsDenormalizedBeforeUpdateHooks = append(requestsDenormalizedBeforeUpdateHooks, requestsDenormalizedHook)
	case boil.AfterUpdateHook:
		requestsDenormalizedAfterUpdateHooks = append(requestsDenormalizedAfterUpdateHooks, requestsDenormalizedHook)
	case boil.BeforeDeleteHook:
		requestsDenormalizedBeforeDeleteHooks = append(requestsDenormalizedBeforeDeleteHooks, requestsDenormalizedHook)
	case boil.AfterDeleteHook:
		requestsDenormalizedAfterDeleteHooks = append(requestsDenormalizedAfterDeleteHooks, requestsDenormalizedHook)
	case boil.BeforeUpsertHook:
		requestsDenormalizedBeforeUpsertHooks = append(requestsDenormalizedBeforeUpsertHooks, requestsDenormalizedHook)
	case boil.AfterUpsertHook:
		requestsDenormalizedAfterUpsertHooks = append(requestsDenormalizedAfterUpsertHooks, requestsDenormalizedHook)
	}
}

// One returns a single requestsDenormalized record from the query.
func (q requestsDenormalizedQuery) One(ctx context.Context, exec boil.ContextExecutor) (*RequestsDenormalized, error) {
	o := &RequestsDenormalized{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for requests_denormalized")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all RequestsDenormalized records from the query.
func (q requestsDenormalizedQuery) All(ctx context.Context, exec boil.ContextExecutor) (RequestsDenormalizedSlice, error) {
	var o []*RequestsDenormalized

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to RequestsDenormalized slice")
	}

	if len(requestsDenormalizedAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all RequestsDenormalized records in the query.
func (q requestsDenormalizedQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count requests_denormalized rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q requestsDenormalizedQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if requests_denormalized exists")
	}

	return count > 0, nil
}

// RequestsDenormalizeds retrieves all the records using an executor.
func RequestsDenormalizeds(mods ...qm.QueryMod) requestsDenormalizedQuery {
	mods = append(mods, qm.From("\"requests_denormalized\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"requests_denormalized\".*"})
	}

	return requestsDenormalizedQuery{q}
}

// FindRequestsDenormalized retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindRequestsDenormalized(ctx context.Context, exec boil.ContextExecutor, iD int, timestamp time.Time, selectCols ...string) (*RequestsDenormalized, error) {
	requestsDenormalizedObj := &RequestsDenormalized{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"requests_denormalized\" where \"id\"=$1 AND \"timestamp\"=$2", sel,
	)

	q := queries.Raw(query, iD, timestamp)

	err := q.Bind(ctx, exec, requestsDenormalizedObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from requests_denormalized")
	}

	if err = requestsDenormalizedObj.doAfterSelectHooks(ctx, exec); err != nil {
		return requestsDenormalizedObj, err
	}

	return requestsDenormalizedObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *RequestsDenormalized) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no requests_denormalized provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(requestsDenormalizedColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	requestsDenormalizedInsertCacheMut.RLock()
	cache, cached := requestsDenormalizedInsertCache[key]
	requestsDenormalizedInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			requestsDenormalizedAllColumns,
			requestsDenormalizedColumnsWithDefault,
			requestsDenormalizedColumnsWithoutDefault,
			nzDefaults,
		)
		wl = strmangle.SetComplement(wl, requestsDenormalizedGeneratedColumns)

		cache.valueMapping, err = queries.BindMapping(requestsDenormalizedType, requestsDenormalizedMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(requestsDenormalizedType, requestsDenormalizedMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"requests_denormalized\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"requests_denormalized\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into requests_denormalized")
	}

	if !cached {
		requestsDenormalizedInsertCacheMut.Lock()
		requestsDenormalizedInsertCache[key] = cache
		requestsDenormalizedInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the RequestsDenormalized.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *RequestsDenormalized) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	requestsDenormalizedUpdateCacheMut.RLock()
	cache, cached := requestsDenormalizedUpdateCache[key]
	requestsDenormalizedUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			requestsDenormalizedAllColumns,
			requestsDenormalizedPrimaryKeyColumns,
		)
		wl = strmangle.SetComplement(wl, requestsDenormalizedGeneratedColumns)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update requests_denormalized, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"requests_denormalized\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, requestsDenormalizedPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(requestsDenormalizedType, requestsDenormalizedMapping, append(wl, requestsDenormalizedPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update requests_denormalized row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for requests_denormalized")
	}

	if !cached {
		requestsDenormalizedUpdateCacheMut.Lock()
		requestsDenormalizedUpdateCache[key] = cache
		requestsDenormalizedUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q requestsDenormalizedQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for requests_denormalized")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for requests_denormalized")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o RequestsDenormalizedSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), requestsDenormalizedPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"requests_denormalized\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, requestsDenormalizedPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in requestsDenormalized slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all requestsDenormalized")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *RequestsDenormalized) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no requests_denormalized provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(requestsDenormalizedColumnsWithDefault, o)

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

	requestsDenormalizedUpsertCacheMut.RLock()
	cache, cached := requestsDenormalizedUpsertCache[key]
	requestsDenormalizedUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			requestsDenormalizedAllColumns,
			requestsDenormalizedColumnsWithDefault,
			requestsDenormalizedColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			requestsDenormalizedAllColumns,
			requestsDenormalizedPrimaryKeyColumns,
		)

		insert = strmangle.SetComplement(insert, requestsDenormalizedGeneratedColumns)
		update = strmangle.SetComplement(update, requestsDenormalizedGeneratedColumns)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert requests_denormalized, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(requestsDenormalizedPrimaryKeyColumns))
			copy(conflict, requestsDenormalizedPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"requests_denormalized\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(requestsDenormalizedType, requestsDenormalizedMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(requestsDenormalizedType, requestsDenormalizedMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert requests_denormalized")
	}

	if !cached {
		requestsDenormalizedUpsertCacheMut.Lock()
		requestsDenormalizedUpsertCache[key] = cache
		requestsDenormalizedUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single RequestsDenormalized record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *RequestsDenormalized) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no RequestsDenormalized provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), requestsDenormalizedPrimaryKeyMapping)
	sql := "DELETE FROM \"requests_denormalized\" WHERE \"id\"=$1 AND \"timestamp\"=$2"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from requests_denormalized")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for requests_denormalized")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q requestsDenormalizedQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no requestsDenormalizedQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from requests_denormalized")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for requests_denormalized")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o RequestsDenormalizedSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(requestsDenormalizedBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), requestsDenormalizedPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"requests_denormalized\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, requestsDenormalizedPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from requestsDenormalized slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for requests_denormalized")
	}

	if len(requestsDenormalizedAfterDeleteHooks) != 0 {
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
func (o *RequestsDenormalized) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindRequestsDenormalized(ctx, exec, o.ID, o.Timestamp)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *RequestsDenormalizedSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := RequestsDenormalizedSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), requestsDenormalizedPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"requests_denormalized\".* FROM \"requests_denormalized\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, requestsDenormalizedPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in RequestsDenormalizedSlice")
	}

	*o = slice

	return nil
}

// RequestsDenormalizedExists checks if the RequestsDenormalized row exists.
func RequestsDenormalizedExists(ctx context.Context, exec boil.ContextExecutor, iD int, timestamp time.Time) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"requests_denormalized\" where \"id\"=$1 AND \"timestamp\"=$2 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD, timestamp)
	}
	row := exec.QueryRowContext(ctx, sql, iD, timestamp)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if requests_denormalized exists")
	}

	return exists, nil
}
