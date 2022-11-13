package djson

var (
	testCase = `
conf = {
    "secretId": "xxxxxxxxxxxxxxxxx",
    "secretKey": "xxxxxxxxxxxxxxxxxxxxxxxxx",
    "region": "ap-chengdu",
}

dts1 = {
    "secretId": conf.secretId,

    # secretKey is the required but not set here, read it from config
    "secretKey": conf.secretKey, 

    "region": conf.region,
    "name": "dts1",
    "migrateAllTables": true,
    "migrateTables": _me.migrateAllTables ? [
        "db1.hello",
        "db1.hello1",
    ] : null
}
    `
)
