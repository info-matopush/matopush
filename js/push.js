self.addEventListener('push', function(evt) {
    var object = evt.data.json();
    var title = 'タイトルなし';
    var body = '';
    var content = '';
    var icon = '/img/news.png';
    var tag = '';
    var endpoint = '';
    if ('SiteTitle' in object) {
        title = object.SiteTitle;
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
    if ('SiteUrl' in object) {
        tag = object.SiteUrl;
    }
    if ('Endpoint' in object) {
        endpoint = object.Endpoint;
    }

    if (body !== '') {
        evt.waitUntil(
            self.registration.showNotification(
                title,
                {
                    body:    body,
                    data:    {
                        url:       content,
                        endpoint:  endpoint,
                    },
                    icon:    icon,
                    tag:     tag,
                }
            )
        )
    }
});

self.addEventListener('notificationclick', function(evt) {
    var url = evt.notification.data.url;
    evt.notification.close();

    var data = new FormData();
    data.append('endpoint', evt.notification.data.endpoint);
    data.append('url', url);

    // URLが指定されていれば遷移する
    if (url !== "") {
        fetch('api/log', {
            method: 'post',
            body:   data,
        });
        return clients.openWindow(url);
    }
});
