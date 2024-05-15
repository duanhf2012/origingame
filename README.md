```js
use AccDB;
db.RealAreaInfo.insertMany([
	{
		"_id":1,
		"GateList":["127.0.0.1:9001"]
	},
	{
		"_id":2,
		"GateList":["127.0.0.1:9002"]
	}]
);


db.ShowAreaInfo.insertMany([
	{
		 "_id":1,
		 "AreaName":"体验1服",
		 "RealAreaId":1,
		 "ServerMark":0
	},
	{
		 "_id":2,
		 "AreaName":"体验2服(合1)",
		 "RealAreaId":1,
		 "ServerMark":0
	},
	{
		 "_id":3,
		 "AreaName":"体验3服",
		 "RealAreaId":2,
		 "ServerMark":0
	},
])
```

