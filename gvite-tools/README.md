git clone https://github.com/vitelabs/go-vite.git

git checkout [latest release tag]

edit go.mod, replace go-vite

```
go build -i -o vitetools main.go


edit vite.mnemonic
edit received.csv

./vitetools batchSend --tokenId tti_3d1ed2b1151ed9bb64d51fee --fromAddr vite_xxxxx --prePrint true
./vitetools batchSend --tokenId tti_3d1ed2b1151ed9bb64d51fee --fromAddr vite_xxxxx --prePrint false


./vitetools autoReceive --address vite_xxxx
```
