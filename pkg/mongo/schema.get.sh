#!/bin/bash
# 2023-03-05

source ./CONFIG
echo $M_URI
echo $M_COLLECTION

# npm config set registry "http://172.31.252.188:8081/repository/npm-proxy/"
# npm -g install extract-mongo-schema

for mc in $M_COLLECTION
do
    echo $mc
    extract-mongo-schema -d $M_URI -c $mc -o $mc.schema.json
    cp $mc.schema.json $mc.schema.bson
    sed -i "s/$mc/properties/g" $mc.schema.bson
    sed -i "s/{/bson.M{/g" $mc.schema.bson
    sed -i 's/},/}/g' $mc.schema.bson
    sed -i 's/}/},/g' $mc.schema.bson
    sed -i 's/true/true,/g' $mc.schema.bson
    sed -i 's/type/bsonType/g' $mc.schema.bson
    # sed -i ':a;N;$!ba;s/}\n/},\n/g' $mc.schema.bson
    sed -i '0,/bson.M{/s//"$jsonSchema": bson.M{\n    "bsonType": "object",/' $mc.schema.bson
    echo "const "$mc"_collection_name = \""$mc"\"" > $mc.schema.go
    echo "var "$mc"_schema = bson.M{"  >> $mc.schema.go
    cat $mc.schema.bson >> $mc.schema.go
    echo "" >> $mc.schema.go
    echo "}" >> $mc.schema.go

    builder="var "$mc"_schema_builder = mongobuilder.NewQueryBuilder("$mc"_collection_name, $mc"_schema", true)"
    echo $builder >> $mc.schema.go
    rm $mc.schema.json
    rm $mc.schema.bson
done
