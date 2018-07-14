package endpoint

import (
	"encoding/base64"
	"errors"
	"hash/fnv"
	"time"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// Endpoint は通知先識別情報(論理モデル)
type Endpoint struct {
	Endpoint string `json:"endpoint"`
	P256dh   []byte `json:"p256dh"`
	Auth     []byte `json:"auth"`
}

// 通知先情報(物理モデル)
type physicalEndpointInfo struct {
	// Key はendpointをハッシュ化したもの
	// endpointをそのまま使うとKeyとして長すぎてmemcacheへの格納が失敗するため。
	Key        string    `datastore:"-" goon:"id"`
	Endpoint   string    `datastore:"endpoint,noindex"`
	P256dh     []byte    `datastore:"p256dh,noindex"`
	Auth       []byte    `datastore:"auth,noindex"`
	CreateDate time.Time `datastore:"create_date,noindex"`
	AccessDate time.Time `datastore:"access_date,noindex"`
	DeleteFlag bool      `datastore:"delete_flag"`
	DeleteDate time.Time `datastore:"delete_date,noindex"`
}

// endpointは長すぎるので、ハッシュを使ってキーを作成する
func endpointToKeyString(endpoint string) string {
	h := fnv.New64a()
	h.Write([]byte(endpoint))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// datastoreよりendpointを取得する
// goonを使用して取得結果はmemcacheに載せるようにする
func getPhysicalEndpointInfo(ctx context.Context, endpoint string) *physicalEndpointInfo {
	g := goon.FromContext(ctx)
	pei := &physicalEndpointInfo{
		Key: endpointToKeyString(endpoint),
	}
	err := g.Get(pei)
	if err != nil {
		return nil
	}
	return pei
}

func getAllEndpointsQuery() *datastore.Query {
	query := datastore.NewQuery("physicalEndpointInfo").Filter("delete_flag=", false)
	return query
}

func getDeletedEndpointsQuery() *datastore.Query {
	query := datastore.NewQuery("physicalEndpointInfo").Filter("delete_flag=", true)
	return query
}

// Touch はアクセス日時の更新を行う
func (e *Endpoint) Touch(ctx context.Context) error {
	pei := getPhysicalEndpointInfo(ctx, e.Endpoint)
	if pei == nil {
		// 存在しない場合は生成する
		return e.Create(ctx)
	}
	// 存在する場合はアクセス日時を更新する
	pei.AccessDate = time.Now()

	g := goon.FromContext(ctx)
	_, err := g.Put(pei)
	return err
}

// Create はEndpointをDatastoreに保存する
func (e *Endpoint) Create(ctx context.Context) error {
	pei := &physicalEndpointInfo{
		Key:        endpointToKeyString(e.Endpoint),
		Endpoint:   e.Endpoint,
		P256dh:     e.P256dh,
		Auth:       e.Auth,
		CreateDate: time.Now(),
		AccessDate: time.Now(),
	}
	g := goon.FromContext(ctx)
	_, err := g.Put(pei)
	log.Infof(ctx, "Create Endpoint err:%v", err)
	return err
}

// NewFromDatastore はphysicalEndpointからEndpointを作成する
func NewFromDatastore(ctx context.Context, endpoint string) (*Endpoint, error) {
	pei := getPhysicalEndpointInfo(ctx, endpoint)
	if pei == nil {
		return nil, errors.New("not found")
	}
	if pei.DeleteFlag == true {
		return nil, errors.New("endpoint was gone")
	}
	return &Endpoint{
		pei.Endpoint,
		pei.P256dh,
		pei.Auth,
	}, nil
}

// Delete はEndpointを論理削除する
func (e *Endpoint) Delete(ctx context.Context) error {
	pei := &physicalEndpointInfo{
		Key: endpointToKeyString(e.Endpoint),
	}
	g := goon.FromContext(ctx)
	err := g.Get(pei)
	if err != nil {
		return err
	}
	pei.DeleteDate = time.Now()
	pei.DeleteFlag = true
	_, err = g.Put(pei)
	return err
}

// Count はDatastore上のEndpoint数を返却する
func Count(ctx context.Context) int {
	g := goon.FromContext(ctx)
	num, _ := g.Count(getAllEndpointsQuery())
	return num
}

// GetAll はDatastore上のEndpointを全て返却する
func GetAll(ctx context.Context) (dst []Endpoint) {
	g := goon.FromContext(ctx)
	query := getAllEndpointsQuery()
	var list []physicalEndpointInfo
	g.GetAll(query, &list)

	for _, endpoint := range list {
		dst = append(dst, Endpoint{
			endpoint.Endpoint,
			endpoint.P256dh,
			endpoint.Auth,
		})
	}
	log.Debugf(ctx, "有効なendpointの数 %d", len(list))
	return
}

// GetAllDeleted はDatastore上の論理削除されたEndpointを
// 全て取得する
func GetAllDeleted(ctx context.Context) (dst []Endpoint) {
	g := goon.FromContext(ctx)
	query := getDeletedEndpointsQuery()
	var list []physicalEndpointInfo
	g.GetAll(query, &list)

	for _, endpoint := range list {
		dst = append(dst, Endpoint{
			endpoint.Endpoint,
			endpoint.P256dh,
			endpoint.Auth,
		})
	}
	log.Debugf(ctx, "無効なendpointの数 %d", len(list))
	return
}

// Cleanup は論理削除されたデータを物理削除する
func (e *Endpoint) Cleanup(ctx context.Context) {
	g := goon.FromContext(ctx)
	ei := physicalEndpointInfo{Key: endpointToKeyString(e.Endpoint)}
	g.Delete(g.Key(ei))
}
