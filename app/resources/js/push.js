self.addEventListener('push', function(evt) {
    var object = evt.data.json();
    var title = 'タイトルなし';
    var body = '';
    var content = '';
    var icon = '/img/news.png';
    var tag = '';
    var endpoint = '';
    var badge = '';
    var image = '';
    if ('SiteTitle' in object) {
        title = object.SiteTitle;
    }
    if ('SiteIcon' in object) {
        badge = object.SiteIcon;
    }
    if ('ContentTitle' in object) {
        body = object.ContentTitle;
    }
    if ('ContentUrl' in object) {
        content = object.ContentUrl;
    }
    if ('Icon' in object) {
        icon = object.Icon;
    }
    if ('FeedUrl' in object) {
        tag = object.FeedUrl;
    }
    if ('Endpoint' in object) {
        endpoint = object.Endpoint;
    }
    if ('ContentImage' in object) {
        image = object.ContentImage;
    }

    // Endpointに到達したことをログする
    if (content !== "") {
        var data = new FormData();
        data.append('endpoint', endpoint);
        data.append('url', content);
        data.append('command', 'reach');
        fetch('api/log', {
            method: 'post',
            body:   data
        });
    }

    // クライアントにメッセージを送る
    self.clients.matchAll().then(
        clients => clients.forEach(client => client.postMessage({
            'matopush' :  'update'
        }))
    );

    if (body !== '') {
        evt.waitUntil(
            self.registration.showNotification(
                title,
                {
                    body:    body,
                    data:    {
                        url:       content,
                        endpoint:  endpoint
                    },
                    image:   image,
                    icon:    icon,
                    tag:     tag,
                    badge:   badge
                }
            )
        )
    }
});

self.addEventListener('notificationclick', function(evt) {
    var url = evt.notification.data.url;
    var endpoint = evt.notification.data.endpoint;
    evt.notification.close();

    // URLが指定されていれば遷移する
    if (url !== "") {
        // Push通知から該当ページに遷移したことをログする
        var data = new FormData();
        data.append('endpoint', endpoint);
        data.append('url', url);
        data.append('command', 'click');
        fetch('api/log', {
            method: 'post',
            body:   data
        });
        return clients.openWindow(url);
    }
});
