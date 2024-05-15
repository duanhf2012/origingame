```js
use AccDB;
db.RealAreaInfo.createIndex({RealAreaId:1}, {unique:true});
db.RealAreaInfo.insertMany([
	{
		"RealAreaId":1,
		"GateList":["127.0.0.1:9001"]
	},
	{
		"RealAreaId":2,
		"GateList":["127.0.0.1:9002"]
	}]
);


db.ShowAreaInfo.createIndex({ShowAreaId:1}, {unique:true})
db.ShowAreaInfo.insertMany([
	{
		 "ShowAreaId":1,
		 "AreaName":"体验1服",
		 "RealAreaId":1,
		 "ServerMark":0
	},
	{
		 "ShowAreaId":2,
		 "AreaName":"体验2服(合1)",
		 "RealAreaId":1,
		 "ServerMark":0
	},
	{
		 "ShowAreaId":3,
		 "AreaName":"体验3服",
		 "RealAreaId":2,
		 "ServerMark":0
	},
])
```

