package endpoint

import (
	"encoding/base64"
	"errors"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"hash/fnv"
	"time"
)

// 通知先識別情報
type EndpointInfo struct {
	Endpoint string `json:"endpoint"`
	P256dh   []byte `json:"p256dh"`
	Auth     []byte `json:"auth"`
}

type physicalEndpointInfo struct {
	// Uid はendpointをHash化したもの(EndpointInfoではEndpointがidだったため、memcacheへの格納が失敗していた。
	Key        string    `datastore:"-"           goon:"id"`
	Endpoint   string    `datastore:"endpoint,    noindex"`
	P256dh     []byte    `datastore:"p256dh,      noindex"`
	Auth       []byte    `datastore:"auth,        noindex"`
	CreateDate time.Time `datastore:"create_date, noindex"`
	AccessDate time.Time `datastore:"access_date, noindex"`
	DeleteFlag bool      `datastore:"delete_flag"`
	DeleteDate time.Time `datastore:"delete_date, noindex"`
}

// endpointは文字列として長すぎるので、ハッシュを使ってキーを作成する
func endpointToKeyString(endpoint string) string {
	h := fnv.New64a()
	h.Write([]byte(endpoint))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func getPhysicalEndpointInfo(ctx context.Context, endpoint string) *physicalEndpointInfo {
	pei := &physicalEndpointInfo{
		Key: endpointToKeyString(endpoint),
	}
	g := goon.FromContext(ctx)
	err := g.Get(pei)
	if err != nil {
		return nil
	}
	return pei
}

func getAllEndpointQuery() *datastore.Query {
	query := datastore.NewQuery("physicalEndpointInfo").Filter("delete_flag=", false)
	return query
}

func Touch(ctx context.Context, endpointInfo *EndpointInfo) error {
	pei := getPhysicalEndpointInfo(ctx, endpointInfo.Endpoint)
	if pei == nil {
		// 存在しない場合は生成する
		return Create(ctx, endpointInfo)
	}
	// 存在する場合はアクセス日時を記録する
	pei.AccessDate = time.Now()

	g := goon.FromContext(ctx)
	_, err := g.Put(pei)
	return err
}

func Create(ctx context.Context, endpointInfo *EndpointInfo) error {
	pei := &physicalEndpointInfo{
		Key:        endpointToKeyString(endpointInfo.Endpoint),
		Endpoint:   endpointInfo.Endpoint,
		P256dh:     endpointInfo.P256dh,
		Auth:       endpointInfo.Auth,
		CreateDate: time.Now(),
		AccessDate: time.Now(),
	}
	g := goon.FromContext(ctx)
	_, err := g.Put(pei)
	return err
}

func Get(ctx context.Context, endpoint string) (*EndpointInfo, error) {
	pei := getPhysicalEndpointInfo(ctx, endpoint)
	if pei == nil {
		return nil, errors.New("not found.")
	}
	if pei.DeleteFlag == true {
		return nil, errors.New("endpoint was gone.")
	}
	return &EndpointInfo{
		Endpoint: pei.Endpoint,
		Auth:     pei.Auth,
		P256dh:   pei.P256dh,
	}, nil
}

func Delete(ctx context.Context, endpoint string) error {
	pei := &physicalEndpointInfo{
		Key: endpointToKeyString(endpoint),
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

func Count(ctx context.Context) int {
	g := goon.FromContext(ctx)
	num, _ := g.Count(getAllEndpointQuery())
	return num
}

func GetAll(ctx context.Context, dst []EndpointInfo) {
	g := goon.FromContext(ctx)
	query := getAllEndpointQuery()
	var list []physicalEndpointInfo
	g.GetAll(query, &list)

	for _, endpoint := range list {
		dst = append(dst, EndpointInfo{
			Endpoint: endpoint.Endpoint,
			Auth:     endpoint.Auth,
			P256dh:   endpoint.P256dh,
		})
	}
}
