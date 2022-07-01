# fabric部署


## 环境  

| 端口服务               | 类型           | IP        |
| ---------------------- | -------------- | --------- |
| orderer.example.com    | fabric-orderer | 127.0.0.1 |
| peer0.org1.example.com | fabric-peer    | 127.0.0.1 |
| peer0.org2.example.com | fabric-peer    | 127.0.0.1 |

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



`CORE_PEER_ADDRESS`=peer0.org2.example.com:9051  

`CORE_PEER_CHAINCODELISTENADDRESS`=peer0.org2.example.com:9052

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
CHANNEL_NAME=chan
export PATH=../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/configtx

cryptogen generate --config=./organizations/cryptogen/crypto-config-orderer.yaml --output="organizations"
cryptogen generate --config=./organizations/cryptogen/crypto-config-org1.yaml --output="organizations"
cryptogen generate --config=./organizations/cryptogen/crypto-config-org2.yaml --output="organizations"
./organizations/ccp-generate.sh

configtxgen -profile TwoOrgsApplicationGenesis -outputBlock ./channel-artifacts/$CHANNEL_NAME.block -channelID $CHANNEL_NAME
```

### 启动所有服务
```zsh
BLOCKFILE="./channel-artifacts/${CHANNEL_NAME}.block"

DOCKER_SOCK=/var/run/docker.sock docker-compose -f compose/compose-test-net.yaml -f compose/docker/docker-compose-test-net.yaml up -d
```

### 加入通道
```zsh
# /opt/gopath/src/github.com/hyperledger/fabric/peer
ORDERER_CA=./organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem
ORDERER_ADMIN_TLS_SIGN_CERT=./organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
export ORDERER_ADMIN_TLS_PRIVATE_KEY=./organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key

osnadmin channel join --channelID $CHANNEL_NAME --config-block ./channel-artifacts/${CHANNEL_NAME}.block -o localhost:7053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY"

FABRIC_CFG_PATH=$PWD/../config/

PEER0_ORG1_CA=./organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer channel join -b $BLOCKFILE

PEER0_ORG2_CA=./organizations/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

peer channel join -b $BLOCKFILE

# 添加锚节点
docker exec cli ./scripts/setAnchorPeer.sh 1 $CHANNEL_NAME
docker exec cli ./scripts/setAnchorPeer.sh 2 $CHANNEL_NAME
```

### 部署链码

```zsh
peer lifecycle chaincode package basic.tar.gz --path ./chaincode --lang golang --label basic

PEER0_ORG1_CA=./organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer lifecycle chaincode install basic.tar.gz
# 拿到Package ID
peer lifecycle chaincode queryinstalled

PEER0_ORG2_CA=./organizations/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

peer lifecycle chaincode install basic.tar.gz
# 拿到Package ID
peer lifecycle chaincode queryinstalled
# 输出
Installed chaincodes on peer:
Package ID: basic:3122010651424553510596788276e0af4bc2188219e06515a237a27dc86506b6, Label: basic
```

### 批准链码

```zsh
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" --channelID $CHANNEL_NAME --name basic -v 1 --package-id basic:3122010651424553510596788276e0af4bc2188219e06515a237a27dc86506b6 --sequence 1

peer lifecycle chaincode checkcommitreadiness --channelID $CHANNEL_NAME --name basic --version 1 --sequence 1 --output json
```

## 简易步骤  

```zsh
./network.sh up createChannel
./network.sh deployCC -ccn basic -ccp ./chaincodenew -ccl go

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true

PEER0_ORG1_CA=./organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA

export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "./organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"InitAll","Args":[]}'

peer chaincode query -C mychannel -n basic -c '{"Args":["QueryAccountList",""]}'

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "./organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"Args":["CreateRealEstate","5feceb66ffc8","6b86b273ff34","1.002E+02","8.06E+01"]}'

peer chaincode query -C mychannel -n basic -c '{"Args":["QueryRealEstateList","",""]}'
```

---

## 后续升级链码  

```zsh
peer lifecycle chaincode package basic_2.tar.gz --path ./chaincode2 --lang golang --label basic_2.0

export CORE_PEER_TLS_ENABLED=true
# 切环境四连
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer lifecycle chaincode install basic_2.tar.gz

peer lifecycle chaincode queryinstalled
basic_2.0:504ebc1bcdfaa78a46a129a0caa9ae43dc8e7ad716b85538ab323af31fa30661

# 复制package_id
export NEW_CC_PACKAGE_ID=basic_2.0:504ebc1bcdfaa78a46a129a0caa9ae43dc8e7ad716b85538ab323af31fa30661

# 同意
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name basic --version 2.0 --package-id $NEW_CC_PACKAGE_ID --sequence 2 --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

# 切环境四连
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

peer lifecycle chaincode install basic_2.tar.gz

# 同意
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name basic --version 2.0 --package-id $NEW_CC_PACKAGE_ID --sequence 2 --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

# 检查同意情况
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name basic --version 2.0 --sequence 2 --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" --output json

# 提交
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name basic --version 2.0 --sequence 2 --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
```