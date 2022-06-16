# fabric部署


## 环境  

| 端口服务               | 类型           | IP            |
| ---------------------- | -------------- | ------------- |
| orderer.example.com    | fabric-orderer | 192.9.200.125 |
| peer0.org1.example.com | fabric-peer    | 192.9.200.172 |
| peer0.org2.example.com | fabric-peer    | 192.9.200.230 |



## orderer  

### 端口开放  

`ORDERER_GENERAL_LISTENPORT`=7050  

### 挂载  

`./channel-artifacts/genesis.block`  :  `/var/hyperledger/orderer/genesis.block`  
`./crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp`  :  `/var/hyperledger/orderer/msp`  
`./crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/`  :  `/var/hyperledger/orderer/tls`  



## peer    

`CORE_PEER_ADDRESS`=peer0.org1.example.com:7051  

`CORE_PEER_CHAINCODELISTENADDRESS`=peer0.org1.example.com:7052

`/var/run/`  :  `/host/var/run/`  
`./crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp`  :  `/etc/hyperledger/fabric/msp`  

`./crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls`  :  `/etc/hyperledger/fabric/tls`



`CORE_PEER_ADDRESS`=peer0.org2.example.com:7051  

`CORE_PEER_CHAINCODELISTENADDRESS`=peer0.org2.example.com:7052

`/var/run/`  :  `/host/var/run/`  
`./crypto-config/peerOrganizations/org2.example.com/peers/peer0.org1.example.com/msp`  :  `/etc/hyperledger/fabric/msp`  

`./crypto-config/peerOrganizations/org2.example.com/peers/peer0.org1.example.com/tls`  :  `/etc/hyperledger/fabric/tls`



## cli  

### 挂载点  

`/var/run/`  :  `/host/var/run/`  

`./chaincode/go/`  :  `/opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/go`

`./crypto-config`  :  `/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/`

`./channel-artifacts`  :  `/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts`

---

## 部署步骤  
### 管理者创建证书  

```zsh


./bin/cryptogen generate --config=./crypto-config.yaml
mkdir channel-artifacts
./bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block -channelID system
./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/mychannel.tx -channelID mychannel

# 复制到192.9.200.172
scp -r raft/resource/channel-artifacts bruce@192.9.200.172:/home/bruce/raft/resource/
scp -r raft/resource/channel-artifacts bruce@192.9.200.232:/home/bruce/raft/resource/
```

### 启动所有服务
```zsh
# 192.9.200.125
docker-compose -f docker-compose-orderer.yaml up -d
# 192.9.200.172
docker-compose -f docker-compose-peer0-org1.yaml up -d
# 192.9.200.232
docker-compose -f docker-compose-peer0-org2.yaml up -d
```

### 加入通道
```zsh
# 192.9.200.172
# 下载链码依赖
cd chaincode/go/abstore
go mod vendor
cd ../../..

docker exec -it cli bash
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/mychannel.tx --outputBlock ./channel-artifacts/mychannel.block --tls --cafile $ORDERER_CA
# peer 加入channel
peer channel join -b ./channel-artifacts/mychannel.block

#成功后复制文件
scp -r ./channel-artifacts bruce@192.9.200.232:/home/bruce/raft/resource

# 192.9.200.232
# 下载链码依赖
cd chaincode/go/abstore
go mod vendor
cd ../../..

docker exec -it cli bash
# peer 加入channel
peer channel join -b ./channel-artifacts/mychannel.block
```

### 部署链码

```zsh
# 192.9.200.172
docker exec -it cli bash
peer lifecycle chaincode package basic.tar.gz --path ./chaincode/go/abstore --lang golang --label basic
peer lifecycle chaincode install basic.tar.gz

# 192.9.200.232
docker exec -it cli bash
peer lifecycle chaincode package basic.tar.gz --path ./chaincode/go/abstore --lang golang --label basic
peer lifecycle chaincode install basic.tar.gz
```

### 批准链码

```zsh
# 拿到Package ID
peer lifecycle chaincode queryinstalled

# 输出
Installed chaincodes on peer:
Package ID: basic:6f292c790b756d2cb8a35de6c421187b533b9917d56458189e12e09fe34984cd, Label: basic


# 192.9.200.172
docker exec -it cli bash
# 查看链码批准情况
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name basic -v 1 --sequence 1 --output json --init-required


ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg --tls true --cafile $ORDERER_CA --channelID mychannel -n basic -v 1 --init-required --package-id basic:6f292c790b756d2cb8a35de6c421187b533b9917d56458189e12e09fe34984cd --sequence 1 --waitForEvent


# 查看链码批准情况
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name basic -v 1 --sequence 1 --output json --init-required

# 192.9.200.232
# 查看链码批准情况
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name basic -v 1 --sequence 1 --output json --init-required

ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg --tls true --cafile $ORDERER_CA --channelID mychannel -n basic -v 1 --init-required --package-id basic:6f292c790b756d2cb8a35de6c421187b533b9917d56458189e12e09fe34984cd --sequence 1 --waitForEvent

# 查看链码批准情况
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name basic -v 1 --sequence 1 --output json --init-required
```

### 批准后提交链码  

```zsh
# 192.9.200.172(任意一台peer)
# 一定要确保每个已批准的组织同时提交，即每个已批准的组织内的peer一同参与提交
peer lifecycle chaincode commit -o orderer.example.com:7050 --tls true --cafile $ORDERER_CA -C mychannel -n basic -v 1 --sequence 1 --init-required --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

# 查看提交情况
peer lifecycle chaincode querycommitted -C mychannel -n basic

Committed chaincode definition for chaincode 'basic' on channel 'mychannel':
Version: 1, Sequence: 1, Endorsement Plugin: escc, Validation Plugin: vscc, Approvals: [Org1MSP: true, Org2MSP: true]
```

### 初始化合约  

```zsh
# 192.9.200.172(任意一台peer)
# 一定要确保每个已批准的组织同时参与到初始化，即每个已批准的组织内的peer一同参与初始化
peer chaincode invoke -o orderer.example.com:7050 --tls true --cafile $ORDERER_CA -C mychannel -n basic --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt  --isInit -c '{"Args":["Init","a","100","b","100"]}'

peer chaincode query -C mychannel -n basic -c '{"Args":["query","a"]}'

# 输出
100
```

