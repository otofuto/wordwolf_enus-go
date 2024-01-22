var CACHE_NAME = "wordwolf-tokyo-20240120";
var urlsToCache = [
	"/",
    "/howto",
    "/offsetting",
    "/name",
    "/game",
    "/play",
    "/finish",
    "/announce",
    "/allquestions",
    "/about",
    "/odai",
    "/questions",
    "/static/css/styles.css",
    "/static/css/master.css",
    "/static/js/scripts_off.js",
    "/static/privacypolicy.txt",
    "/static/role.txt",
];

self.addEventListener('install', function(event) {
    event.waitUntil(
        caches
        .open(CACHE_NAME)
        .then(function(cache){
            return cache.addAll(urlsToCache);
        })
    );
});

self.addEventListener('fetch', function(e) {
    e.respondWith(
        caches.match(e.request)
        .then(res => {
            return res || fetch(e.request);
        })
    );
});

self.addEventListener('activate', function(event) { 
    event.waitUntil(
        caches.keys().then(function(cacheNames) {
            return Promise.all(
                cacheNames.filter(function(cacheName) {
                    return cacheName !== CACHE_NAME;
                }).map(function(cacheName) {
                    console.info("delete cache: " + cacheName);
                    return caches.delete(cacheName);
                })
            );
        }).then(() => {
            self.clients.claim();
        })
    );
});