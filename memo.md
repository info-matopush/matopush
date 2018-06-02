## manifest.json

PWAとして動作させるために記述が必要となる。

128x128サイズのアイコンはホーム画面からの起動時にスプラッシュ画面に使用される。

-----
## indexedDB
クライアント側で何か保存したいものがある場合に使用する。

- Dexie.js

indexedDBはそのままでは使用しにくいので、ライブラリを使用する。
http://dexie.org/

-----
## データ構造

### テーブル一覧

| テーブル名 | 役割 |
| ---: | :--- |
| Endpoint | エンドポイント=通知先(ブラウザ)を管理する |
| Site | 巡回するサイトの情報を管理する |
| SiteSubscribe | エンドポイントに紐づくサイト情報を管理する |
| Content | サイトを巡回して得たコンテンツ情報を管理する |
| Property | プロパティを管理する |
| ServerKey | サーバのキー情報(鍵ペア)を管理する |

-----
### プロパティ一覧

| プロパティ名 |  |
| ---: | :--- |
| google.custom.search.apikey | googleカスタムサーチAPIを使用するのに必要なキー情報 |
| google.search.engine.id | google検索エンジンのID |

-----
## API

### エンドポイント登録

endpointを登録する。

#### インターフェース

| 属性 | 値 | デフォルト |
| --- | --- | --- |
| パス | /api/regist | |
| パラメータ(必須) | endpoint | |
| パラメータ(必須) | p256dh | |
| パラメータ(必須) | auth | |

#### シーケンス

```mermaid
sequenceDiagram
    participant A as ブラウザ
    participant B as サーバ
    participant C as Datastore(Endpoint)
    
    A->>B: /api/conf/site
    B->>C: データ取得
    C-->>B: データ応答
    opt 失敗(新規登録)
        B->>C: データ登録
    end
    B->>C: データ更新
```


-----
### エンドポイント解除

endpointを解除する。

#### インターフェース

| 属性 | 値 | デフォルト |
| --- | --- | --- |
| パス | /api/unregist | |
| パラメータ(必須) | endpoint | |

#### シーケンス

```mermaid
sequenceDiagram
    participant A as ブラウザ
    participant B as サーバ
    participant C as Datastore(Endpoint)
    
    A->>B: /api/unregist
    B->>C: データ削除(deleteフラグ=true)
    B-->>A: 結果(bodyなし)
```


-----
### 購読サイト設定

endpointに紐づく購読中のサイト情報を設定する。

#### インターフェース

| 属性 | 値 | デフォルト |
| --- | --- | --- |
| パス | /api/conf/site | |
| パラメータ(必須) | endpoint | |
| パラメータ(必須) | siteUrl | |
| パラメータ(オプション) | value | false |

#### シーケンス

```mermaid
sequenceDiagram
    participant A as ブラウザ
    participant B as サーバ
    participant C as Datastore(Site)
    participant D as Datastore(SiteSubscribe)
    participant E as 他サイト(siteUrl)
    
    A->>B: /api/conf/site
    B->>C: データ取得
    C-->>B: データ応答
    opt データ取得失敗(新規登録)
        B->>E: サイトデータ取得
        B->>C: サイトデータ登録
    end
    B->>D: 購読情報更新
    B-->>A: 結果返却
```

-----
### 購読サイトリスト取得

endpointに紐づく購読中のサイト情報を全て返却する。

#### インターフェース

| 属性 | 値 |
| --- | --- |
| パス | /api/conf/list |
| パラメータ(必須) | endpoint |

#### シーケンス

```mermaid
sequenceDiagram
    participant A as ブラウザ
    participant B as サーバ
    participant C as Datastore(SiteSubscribe)
    participant D as Datastore(Site)
    
    A->>B: /api/conf/list
    B->>C: リスト取得
    Note right of C: datastore.Queryを<br/>用いて購読情報を<br/>全部取得
    C-->>B: リスト応答
    loop 取得件数 
    B->>D: データ取得
    Note right of D: datastore.Getを<br/>用いてサイト情報を<br/>取得
    D-->>B: データ応答
    end
    B-->>A: リストを返却
```
-----
### 購読サイト削除

購読情報から指定されたサイトを削除する。

#### インターフェース

| 属性 | 値 |
| --- | --- |
| パス | /api/conf/remove |
| パラメータ(必須) | endpoint |
| パラメータ(必須) | feedUrl |

#### シーケンス

```mermaid
sequenceDiagram
    participant A as ブラウザ
    participant B as サーバ
    participant C as Datastore(SiteSubscribe)
    
    A->>B: /api/conf/remove
    B->>C: データ更新
    B-->>A: 結果(bodyなし)
```
-----
## cron

### サイト更新情報通知

サイトを巡回し、更新情報があれば登録されたendpointに通知する。

#### インターフェース

| 属性 | 値 |
| --- | --- |
| パス | /admin/api/cron |

-----
### ヘルスチェック

endpointに不可視の通知を行い、無効なendpointを検出する。

#### インターフェース

| 属性 | 値 |
| --- | --- |
| パス | /admin/api/health |

-----
### サイト情報クリーンナップ

不要な情報を削除する。

#### インターフェース

| 属性 | 値 |
| --- | --- |
| パス | /admin/api/cleanup |

-----
## TaskQueue

### サイト更新通知

サイトのFeed情報を読み込み、更新を検知した場合はWebPushで購読しているエンドポイントへ通知を行う。

#### インターフェース

| 属性 | 値 |
| --- | --- |
| パス | /admin/api/publish |
| パラメータ(必須) | FeedURL |


-----
## markdown

参考サイト

https://mermaidjs.github.io/sequenceDiagram.html
