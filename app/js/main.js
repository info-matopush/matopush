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

var tabIndex = 0;

var db = new Dexie("key_value_database");
db.version(1).stores({
    keyvalue: 'key, value'
});

window.addEventListener('load', function() {
    var vue = new Vue({
        el: '#matopush',
        data: {
            searchList,
            myList,
            publicList,
            timelineList,
            tabIndex
        },
        methods: {
            openUrl: function (url) {
                window.open(url);
            },
            addfeed: function (index) {
                var sel = publicList.items[index];
                sel.Value = true;
                myList.items.push(sel);
                publicList.items.splice(index, 1);
                toggleSubscribe(sel);
            },
            addSite: function (feedURL) {
                var sendData = new FormData();
                sendData.append('endpoint', subscription.endpoint);
                sendData.append('siteUrl', feedURL);
                sendData.append('value', 'true')

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
            },
            remove: function (index) {
                var sel = myList.items[index];
                var sendData = new FormData();
                sendData.append('endpoint', subscription.endpoint);
                sendData.append('feedUrl', sel.FeedUrl);

                $.ajax({
                    url: "/api/conf/remove",
                    type: "POST",
                    data: sendData,
                    processData: false,
                    contentType: false,
                    success:
                        function (resp) {
                            myList.items.splice(index, 1);
                        },
                });
            },
            toggleAtMyList: function (index) {
                var sel = myList.items[index];
                toggleSubscribe(sel);
            },
            ontabclick: function(index) {
                db.keyvalue.put({key: "latest_tab", value: index});
            },
        }
    });

    navigator.serviceWorker.addEventListener('message', onWorkerMessage);

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

    db.keyvalue.get('latest_tab').then(function(tab) {
        if (tab.value === 'site') {
            vue.tabIndex = 0;
        } else {
            vue.tabIndex = 1;
        }    
    });

}, false);

// ServiceWorkerの通知を取得。
// 画面を更新する
function onWorkerMessage() {
    refreshMyList();
}

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
    myList.items.splice(0);
    timelineList.items.splice(0);
    if (items == null) {
        return
    }
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
            var date = items[i].Contents[j].ModifyDate;
            date = date.replace('T', ' ');
            date = date.replace('+09:00', '');
            items[i].Contents[j].ModifyDate = date;
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
                setMyList(resp);
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

