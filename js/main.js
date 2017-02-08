'use strict';

let _ = function(id) {return document.getElementById(id);}
let registURL = '/api/regist';
let unregistURL = 'api/unregist';
let subscription = null;
let serverKey = null;

window.addEventListener('load', function() {
    fetch('api/list', {
        method: 'get'
    }).then(function(resp) {
        return resp.json();
    }).then(function(json) {
        // 動的に記事リストを作る
        for (var i = 0; i < json.length; i++) {
            let site = json[i].SiteUrl

            // 後で削除できるようにidを振っておく
            let tr = document.createElement('tr');
            tr.id = 'parent_'+site

            let subscribe = document.createElement('input');
            subscribe.id = site;
            subscribe.type = 'checkbox';
            subscribe.checked = false;
            subscribe.value = '購読する';
            subscribe.onclick = function(){toggleSubscribe(site)};

            let contentTitle = document.createElement('a');
            contentTitle.id = 'contentTitle';
            contentTitle.href = json[i].ContentUrl;
            contentTitle.textContent = json[i].ContentTitle;

            let th = document.createElement('th');
            th.scope = 'row';
            th.appendChild(subscribe);
            tr.appendChild(th);
            let td1 = document.createElement('td');
            td1.textContent = json[i].SiteTitle;
            tr.appendChild(td1);

            let td2 = document.createElement('td');
            td2.appendChild(contentTitle);
            tr.appendChild(td2);

            let td3 = document.createElement('td');
            td3.textContent = json[i].SiteUrl;
            tr.appendChild(td3);
            _('siteTable').appendChild(tr);
        }

        // 画面が作られたらWebPushの準備を行う
        if ('serviceWorker' in navigator) {
            _('subscribe').addEventListener('click', togglePushSubscription, false);
            _('test').addEventListener('click', testPush, false);
            _('addSite').addEventListener('click', addSite, false);
            fetch('./api/key').then(getServerKey).then(setServerKey);
            navigator.serviceWorker.register('push.js')
        }
    });
}, false)

function addSite() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    data.append('siteUrl', _('SiteUrl').value)

    fetch('api/add', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.text();
    }).then(function(text) {
        alert(text);
        location.reload();
    });
}

function testPush() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);

    fetch('api/test', {
        method: 'post',
        body: data
    });
    document.activeElement.blur();
}

function toggleSubscribe(key) {
    var data = new FormData();
    if (subscription == null) {
        alert('プッシュ通知が有効になっていません。');
        location.reload();
        return;
    }
    data.append('endpoint', subscription.endpoint);
    data.append('siteUrl', key);
    data.append('value', _(key).checked);

    fetch('api/conf/site', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.text();
    }).then(function(text) {
        alert(text);
        location.reload();
    });
}

function encodeBase64URL(buffer) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(buffer))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function decodeBase64URL(str) {
    let dec = atob(str.replace(/\-/g, '+').replace(/_/g, '/'));
    let buffer = new Uint8Array(dec.length);
    for(let i = 0 ; i < dec.length ; i++)
        buffer[i] = dec.charCodeAt(i);
    return buffer;
}

function decodeBase64URL(str) {
    let dec = atob(str.replace(/\-/g, '+').replace(/_/g, '/'));
    let buffer = new Uint8Array(dec.length);
    for(let i = 0 ; i < dec.length ; i++)
        buffer[i] = dec.charCodeAt(i);
    return buffer;
}

function getServerKey(resp) {
    return resp.text();
}

function setServerKey(key) {
    serverKey = decodeBase64URL(key);
    navigator.serviceWorker.ready.then(serviceWorkerReady)
}

function serviceWorkerReady(registration) {
    if ('pushManager' in registration) {
        registration.pushManager.getSubscription().then(getSubscription);
    }
    else {
        alert('プッシュ通知を有効にできません。')
    }
}

function togglePushSubscription() {
    _('subscribe').disabled = true;

    if (!_('subscribe').classList.contains('subscribing')) {
        requestNotificationPermission();
    }
    else {
        requestPushUnsubscription();
    }
}

function requestNotificationPermission() {
    Notification.requestPermission(function(permission) {
        if (permission !== 'denied') {
            requestPushPermission();
        }
    });
}

function requestPushPermission() {
    if ('permissions' in navigator)
        navigator.permissions.query({
            name: 'push',
            userVisibleOnly: true
        }).then(checkPushPermission);
    else if (Notification.permission !== 'denied') {
        navigator.serviceWorker.ready.then(requestPushSubscription);
    }
}

function checkPushPermission(evt) {
    let state = evt.state || evt.status;
    if (state !== 'denied')
        navigator.serviceWorker.ready.then(requestPushSubscription);
}

function requestPushSubscription(registration) {
    let opt = {
        userVisible: true,
        userVisibleOnly: true,
        applicationServerKey: serverKey
    };
    return registration.pushManager.subscribe(opt).then(getSubscription, errorSubscription);
}

function errorSubscription(err) {
    alert('プッシュ通知を有効にできません。' + err);
}

function getSubscription(sub) {
    if (sub) {
        enablePushRequest(sub);
    }
    else {
        disablePushRequest();
    }
}

function requestPushUnsubscription() {
    if (subscription) {
        subscription.unsubscribe();

        // subscriptionを削除する
        var data = new FormData();
        data.append('endpoint', subscription.endpoint);
        fetch(unregistURL, {
            method: 'post',
            body:   data
        }).then(res => {
        });

        subscription = null;
        disablePushRequest();
        location.reload();
    }
}

function disablePushRequest() {
    _('subscribe').classList.remove('subscribing');
    _('subscribe').disabled = false;
    _('test').disabled = true;
    _('addSite').disabled = true;
}

function enablePushRequest(sub) {
    subscription = sub;
    _('subscribe').classList.add('subscribing');
    _('subscribe').disabled = false;
    _('test').disabled = false;
    _('addSite').disabled = false;

    // subscriptionを登録する
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    data.append('auth',     encodeBase64URL(subscription.getKey('auth')));
    data.append('p256dh',   encodeBase64URL(subscription.getKey('p256dh')));
    fetch(registURL, {
        method: 'post',
        body:   data
    }).then(res => {
    });

    fetch('api/conf/list', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.json();
    }).then(function(json) {
        console.log(json);
        // 動的にマイ・サイトのリストを作る
        for (var i = 0; i < json.length; i++) {
            let site = json[i].SiteUrl;

            let target = _('parent_'+site);
            if (target !== null) {
                target.parentNode.removeChild(target);
            }

            let subscribe = document.createElement('input');
            subscribe.id = site;
            subscribe.type = 'checkbox';
            subscribe.checked = false;
            subscribe.value = '購読する';
            subscribe.onclick = function(){toggleSubscribe(site)};

            let contentTitle = document.createElement('a');
            contentTitle.id = 'contentTitle';
            contentTitle.href = json[i].ContentUrl;
            contentTitle.textContent = json[i].ContentTitle;

            let tr = document.createElement('tr');
            let th = document.createElement('th');
            th.scope = 'row';
            th.appendChild(subscribe);
            tr.appendChild(th);
            let td1 = document.createElement('td');
            td1.textContent = json[i].SiteTitle;
            tr.appendChild(td1);

            let td2 = document.createElement('td');
            td2.appendChild(contentTitle);
            tr.appendChild(td2);

            let td3 = document.createElement('td');
            td3.textContent = json[i].SiteUrl;
            tr.appendChild(td3);
            _('mySiteTable').appendChild(tr);
        }
        // チェックを更新する
        for (var i = 0; i < json.length; i++) {
            let site = json[i].SiteUrl;
            if (_(site) !== null) {
                if (json[i].Value === 'true')
                    _(site).checked = true;
                else
                    _(site).checked = false;
            }
        }
    });
}

