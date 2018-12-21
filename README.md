Pixel - simple requests counter


## Start
```
➜  go get github.com/recoilme/pixel  
➜  go get ./...
➜  go build
➜  ./pixel --port=3000 --debug=true
```

**How to use**

Pixel start weserver with 2 routers:

 - /stats/:group
 - /write/:group/:counter


**Write**

Send Get request on write route:

```
curl -v localhost:3000/write/group/key1
curl -v localhost:3000/write/group/key1
curl -v localhost:3000/write/group/key2
```

Pixel will count get requests. 


Return code:
 - 200 OK
 - 422 and error string in case of error


**Stats**

Send Get request on stats route:

```
curl -v localhost:3000/stats/group
```

Response example:

```
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Fri, 21 Dec 2018 13:12:45 GMT
< Content-Length: 51
<
* Connection #0 to host localhost left intact
[{"group":"key1","hit":2},{"group":"key2","hit":1}]
```

Return code:
 - 200 OK
 - 422 and error string in case of error

**How it work**

Pixel use pudge as key value store. 

 - Group will be a file. 
 - Counter will be a Key 