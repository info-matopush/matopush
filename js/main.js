'use strict';

var _ = function(id) {return document.getElementById(id);};
var registURL = '/api/regist';
var unregistURL = 'api/unregist';
var subscription = null;
var serverKey = null;

var publicList = {
    items:[]
};

var myList = {
    items: []
};

window.addEventListener('load', function() {
    new Vue({
        el: '#my-list',
        data: myList,
        methods: {
            onclick: function (index) {
                var sel = myList.items[index];
                toggleSubscribe(sel);
            }
        }
    });

    new Vue({
        el: '#public-list',
        data: publicList,
        methods: {
            onclick: function (index) {
                var sel = publicList.items[index];
                myList.items.push(sel);
                publicList.items.splice(index, 1);
                toggleSubscribe(sel);
            }
        }
    });

    fetch('api/list', {
        method: 'get'
    }).then(function(resp) {
        return resp.json();
    }).then(function(json) {
        if (json != null) {
            publicList.items = json;
        }

        // 画面が作られたらWebPushの準備を行う
        if ('serviceWorker' in navigator) {
            _('subscribe').addEventListener('click', togglePushSubscription, false);
            _('test').addEventListener('click', testPush, false);
            _('addSite').addEventListener('click', addSite, false);
//            _('searchSite').addEventListener('click', searchSite, false);
            fetch('./api/key').then(getServerKey).then(setServerKey);
            navigator.serviceWorker.register('push.js');
        }
    });
}, false);

function setMyList(items) {
    // 重複する登録済みリストから消し込みを行う
    for (var i=0; i<items.length; i++) {
        for (var j=0; j<publicList.items.length; j++) {
            if (publicList.items[j].FeedUrl === items[i].FeedUrl) {
                publicList.items.splice(j, 1);
            }
        }
    }
    myList.items = items;
}

function toggleSubscribe(item) {
    console.log(item);
    var data = new FormData();
    if (subscription == null) {
        alert('プッシュ通知が有効になっていません。');
        location.reload();
        return;
    }
    data.append('endpoint', subscription.endpoint);
    data.append('siteUrl', item.FeedUrl);
    if (item.Value) {
        data.append('value', "true");
    } else {
        data.append('value', "false");
    }

    fetch('api/conf/site', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.text();
    }).then(function(text) {
        alert(text);
    });
}

var searchResult;

function searchSite() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    data.append('keyword', _('keyword').value);

    fetch('api/search', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.json();
    }).then(function(json) {
        searchResult = json;
    });
}

function addSite() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    data.append('siteUrl', _('siteUrl').value);

    _('siteUrl').value = '';

    fetch('api/add', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.text();
    }).then(function(text) {
        alert(text);
        refreshMyList();
    });
}

function refreshMyList() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    fetch('api/conf/list', {
        method: 'post',
        body: data
    }).then(function(resp) {
        return resp.json();
    }).then(function(json) {
        if (json != null) {
            setMyList(json);
        }
    });
}

function testPush() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);

    fetch('api/test', {
        method: 'post',
        body: data
    });
}

function encodeBase64URL(buffer) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(buffer))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function decodeBase64URL(str) {
    var dec = atob(str.replace(/\-/g, '+').replace(/_/g, '/'));
    var buffer = new Uint8Array(dec.length);
    for(var i = 0 ; i < dec.length ; i++)
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
    var state = evt.state || evt.status;
    if (state !== 'denied')
        navigator.serviceWorker.ready.then(requestPushSubscription);
}

function requestPushSubscription(registration) {
    var opt = {
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
        }).then(function(){
            subscription = null;
            disablePushRequest();
            location.reload();
        });
    }
}

function disablePushRequest() {
    _('subscribe').classList.remove('subscribing');
    _('subscribe').disabled = false;
    _('test').disabled = true;
    _('addSite').disabled = true;
    _('siteUrl').disabled = true;
}

function enablePushRequest(sub) {
    subscription = sub;
    _('subscribe').classList.add('subscribing');
    _('subscribe').disabled = false;
    _('test').disabled = false;
    _('addSite').disabled = false;
    _('siteUrl').disabled = false;

    // subscriptionを登録する
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    data.append('auth',     encodeBase64URL(subscription.getKey('auth')));
    data.append('p256dh',   encodeBase64URL(subscription.getKey('p256dh')));
    fetch(registURL, {
        method: 'post',
        body:   data
    }).then(function(){
        refreshMyList();
    });
}

