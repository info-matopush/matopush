package utility

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type keyValue struct {
	Name        string
	StringValue string `datastore:"str_val,noindex"`
	Int64Value  int64  `datastore:"int_val,noindex"`
}

// Property はデフォルト値を持つプロパティ
// 値はdatastoreに持つ.
type Property struct {
	ctx context.Context
}

// NewFromContext はコンテキストを元にPropertyを作成する
func NewFromContext(ctx context.Context) *Property {
	return &Property{ctx}
}

// GetString は文字列のプロパティ値を取得する
// もしdatastore上にEntityがなければ、デフォルト値を返却するのと同時に
// デフォルト値でEntityを作成する
func (p *Property) GetString(name string, def string) string {
	q := datastore.NewQuery("property").Filter("Name=", name)
	it := q.Run(p.ctx)
	kv := keyValue{name, def, 0}
	_, err := it.Next(&kv)
	if err != datastore.Done {
		return kv.StringValue
	}
	k := datastore.NewIncompleteKey(p.ctx, "property", nil)
	datastore.Put(p.ctx, k, &kv)
	return def
}

// GetInt64 は数値(int64)のプロパティ値を取得する
// もしdatastore上にEntityがなければ、デフォルト値を返却するのと同時に
// デフォルト値でEntityを作成する
func (p *Property) GetInt64(name string, def int64) int64 {
	q := datastore.NewQuery("property").Filter("Name=", name)
	it := q.Run(p.ctx)
	kv := keyValue{name, "", def}
	_, err := it.Next(&kv)
	if err != datastore.Done {
		return kv.Int64Value
	}
	k := datastore.NewIncompleteKey(p.ctx, "property", nil)
	datastore.Put(p.ctx, k, &kv)
	return def
}
