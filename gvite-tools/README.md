```
go build -i -o vitetools *.go


edit vite.mnemonic
edit received.csv

./vitetools batchSend --tokenId tti_3d1ed2b1151ed9bb64d51fee --fromAddr vite_xxxxx --prePrint true
./vitetools batchSend --tokenId tti_3d1ed2b1151ed9bb64d51fee --fromAddr vite_xxxxx --prePrint false


./vitetools autoReceive --address vite_xxxx
```
