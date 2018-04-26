'use strict';

var _ = function(id) {return document.getElementById(id);};
var registURL = '/api/regist';
var unregistURL = '/api/unregist';
var subscription = null;
var serverKey = null;

var publicList = {
    items:[]
};

var myList = {
    items: []
};

var searchList = {
    items: []
};

var timelineList = {
    items: []
};

window.addEventListener('load', function() {
    new Vue({
        el: '#matopush',
        data: {
            searchList,
            myList,
            publicList,
            timelineList,
        },
        methods: {
            openUrl: function (url) {
                window.open(url);
            },
            toggleAtPublicList: function (index) {
                var sel = publicList.items[index];
                myList.items.push(sel);
                publicList.items.splice(index, 1);
                toggleSubscribe(sel);
            },
            addSite: function (feedURL) {
                var sendData = new FormData();
                sendData.append('endpoint', subscription.endpoint);
                sendData.append('siteUrl', feedURL);

                $.ajax({
                    url: "/api/add",
                    type: "POST",
                    data: sendData,
                    processData: false,
                    contentType: false,
                    success:
                        function (resp) {
                            alert(resp);
                            refreshMyList();
                        },
                });
            },
            toggleAtMyList: function (index) {
                var sel = myList.items[index];
                toggleSubscribe(sel);
            }            
        }
    });

    $.ajax({
        url: "/api/list",
        type: "GET",
        dataType: "json",
        processData: false,
        contentData: false,
        success:
            function (resp) {
                if (resp != null) {
                    publicList.items = resp;
                }
                // 画面が作られたらWebPushの準備を行う
                if ('serviceWorker' in navigator) {
                    _('subscribe').addEventListener('click', togglePushSubscription, false);
                    _('test').addEventListener('click', testPush, false);
                    _('addSite').addEventListener('click', addSite, false);
                    _('searchSite').addEventListener('click', searchSite, false);
                    fetch('./api/key').then(getServerKey).then(setServerKey);
                    navigator.serviceWorker.register('push.js');
                }
            },
    });
}, false);

function compare(a, b) {
    const timea = a.ModifyDate;
    const timeb = b.ModifyDate;
    if (timea < timeb) {
        return 1;
    } else if (timea > timeb) {
        return -1;
    }
    return 0;
}

function setMyList(items) {
    timelineList.items.splice(0);
    // 重複する登録済みリストから消し込みを行う
    for (var i=0; i<items.length; i++) {
        // サイト別表示用
        for (var j=0; j<publicList.items.length; j++) {
            if (publicList.items[j].FeedUrl === items[i].FeedUrl) {
                publicList.items.splice(j, 1);
            }
        }
        // タイムライン表示用
        for (var j=0; j<items[i].Contents.length; j++) {
            timelineList.items.push(items[i].Contents[j]);
            timelineList.items.sort(compare)
        }
    }
    myList.items = items;
}

function toggleSubscribe(item) {
    if (subscription == null) {
        alert('プッシュ通知が有効になっていません。');
        location.reload();
        return;
    }

    // 
    var sendData = new FormData();
    sendData.append('endpoint', subscription.endpoint);
    sendData.append('siteUrl', item.FeedUrl);
    if (item.Value) {
        sendData.append('value', "true");
    } else {
        sendData.append('value', "false");
    }

    $.ajax({
        url: "/api/conf/site",
        type: "POST",
        data: sendData,
        processData: false,
        contentType: false,
        success:
            function (resp) {
                alert(resp);
                refreshMyList();
            },
    });
}

function searchSite() {
    var sendData = new FormData();
    sendData.append('endpoint', subscription.endpoint);
    sendData.append('keyword', _('keyword').value);

    $.ajax({
        url: "api/search",
        type: "POST",
        data: sendData,
        dataType: "json",
        processData: false,
        contentType: false,
        success:
            function (resp) {
                searchList.items = resp.items;
            },
    });
}

function addSite() {
    var sendData = new FormData();
    sendData.append('endpoint', subscription.endpoint);
    sendData.append('siteUrl', _('siteUrl').value);

    _('siteUrl').value = '';

    $.ajax({
        url: "/api/add",
        type: "POST",
        data: sendData,
        processData: false,
        contentType: false,
        success:
            function (resp) {
                alert(resp);
                refreshMyList();
            },
    });
}

function refreshMyList() {
    var sendData = new FormData();
    sendData.append('endpoint', subscription.endpoint);

    $.ajax({
        url: "/api/conf/list",
        type: "POST",
        data: sendData,
        dataType: "json",
        processData: false,
        contentType: false,
        success:
            function (resp) {
                if (resp != null) {
                    setMyList(resp);
                }
            },
    });
}

function testPush() {
    var sendData = new FormData();
    sendData.append('endpoint', subscription.endpoint);

    $.ajax({
        url: "/api/test",
        type: "POST",
        data: sendData,
        processData: false,
        contentType: false, 
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
        var sendData = new FormData();
        sendData.append('endpoint', subscription.endpoint);

        $.ajax({
            url: unregistURL,
            type: "POST",
            data: sendData,
            processData: false,
            contentTYpe: false,
            success:
                function () {
                    subscription = null;
                    disablePushRequest();
                    // todo:
                    location.reload();                            
                }
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
    var sendData = new FormData();
    sendData.append('endpoint', subscription.endpoint);
    sendData.append('auth',     encodeBase64URL(subscription.getKey('auth')));
    sendData.append('p256dh',   encodeBase64URL(subscription.getKey('p256dh')));

    $.ajax({
        url: registURL,
        type: "POST",
        data: sendData,
        processData: false,
        contentType: false,
        success:
            function (resp) {
                refreshMyList();
            },
    });
}

