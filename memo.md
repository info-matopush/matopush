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
## API

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
    B->>C: データ削除
    B-->>A: 結果(bodyなし)
```


#### TODO

リスト取得時にkeysOnlyを付与できれば取得コストが下げられる。


## markdown

参考サイト

https://mermaidjs.github.io/sequenceDiagram.html
