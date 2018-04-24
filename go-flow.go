package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/gogap/builder"
	"github.com/gogap/config"
	"github.com/urfave/cli"
)

const (
	flowTempl = "cGFja2FnZSBtYWluCgppbXBvcnQgKAoJImJ5dGVzIgoJImVuY29kaW5nL2Jhc2U2NCIKCSJlbmNvZGluZy9qc29uIgoJImZtdCIKCSJpby9pb3V0aWwiCgkibmV0L3VybCIKCSJvcyIKCSJzdHJjb252IgoJInN0cmluZ3MiCgoJImdpdGh1Yi5jb20vZ29nYXAvY29uZmlnIgoJImdpdGh1Yi5jb20vZ29nYXAvY29udGV4dCIKCSJnaXRodWIuY29tL2dvZ2FwL2Zsb3ciCgkiZ2l0aHViLmNvbS9nb2dhcC9sb2dydXNfbWF0ZSIKCSJnaXRodWIuY29tL3NpcnVwc2VuL2xvZ3J1cyIKCSJnaXRodWIuY29tL3VyZmF2ZS9jbGkiCikKCnZhciAoCgljb25maWdTdHIgPSB7ey5jb25maWdfc3RyfX0KKQoKZnVuYyBpbml0KCkgewoJaWYgbGVuKGNvbmZpZ1N0cikgPiAwIHsKCQljb25maWdTdHJEYXRhLCBlcnIgOj0gYmFzZTY0LlN0ZEVuY29kaW5nLkRlY29kZVN0cmluZyhjb25maWdTdHIpCgkJaWYgZXJyICE9IG5pbCB7CgkJCWVyciA9IGZtdC5FcnJvcmYoImRlY29kZSBjb25maWcgc3RyIGZhaWx1cmU6ICVzXG4iLCBlcnIuRXJyb3IoKSkKCQkJcGFuaWMoZXJyKQoJCX0KCQljb25maWdTdHIgPSBzdHJpbmcoY29uZmlnU3RyRGF0YSkKCX0KfQoKZnVuYyBtYWluKCkgewoKCWNvbmYgOj0gY29uZmlnLk5ld0NvbmZpZygKCQljb25maWcuQ29uZmlnU3RyaW5nKGNvbmZpZ1N0ciksCgkpCgoJYXBwQ29uZiA6PSBjb25mLkdldENvbmZpZygiYXBwIikKCglhcHAgOj0gY2xpLk5ld0FwcCgpCgoJYXBwLlZlcnNpb24gPSBhcHBDb25mLkdldFN0cmluZygidmVyc2lvbiIsICIwLjAuMCIpCglhcHAuQXV0aG9yID0gYXBwQ29uZi5HZXRTdHJpbmcoImF1dGhvciIpCglhcHAuTmFtZSA9IGFwcENvbmYuR2V0U3RyaW5nKCJuYW1lIiwgImFwcCIpCglhcHAuVXNhZ2UgPSBhcHBDb25mLkdldFN0cmluZygidXNhZ2UiKQoJYXBwLkhlbHBOYW1lID0gYXBwLk5hbWUKCglpZiBhcHAuVmVyc2lvbiA9PSAiMC4wLjAiIHsKCQlhcHAuSGlkZVZlcnNpb24gPSB0cnVlCgl9CgoJY29tbWFuZHNDb25mIDo9IGFwcENvbmYuR2V0Q29uZmlnKCJjb21tYW5kcyIpCgoJZm9yIF8sIGtleSA6PSByYW5nZSBjb21tYW5kc0NvbmYuS2V5cygpIHsKCQlnZW5lcmF0ZUNvbW1hbmRzKCZhcHAuQ29tbWFuZHMsIGtleSwgY29tbWFuZHNDb25mLkdldENvbmZpZyhrZXkpKQoJfQoKCWFwcC5SdW5BbmRFeGl0T25FcnJvcigpCgoJcmV0dXJuCn0KCmZ1bmMgbmV3QWN0aW9uKG5hbWUgc3RyaW5nLCBjb25mIGNvbmZpZy5Db25maWd1cmF0aW9uKSBjbGkuQWN0aW9uRnVuYyB7CgoJcmV0dXJuIGZ1bmMoY3R4ICpjbGkuQ29udGV4dCkgKGVyciBlcnJvcikgewoKCQllcnIgPSBsb2FkRU5WKGN0eCkKCgkJaWYgZXJyICE9IG5pbCB7CgkJCXJldHVybgoJCX0KCgkJZGlzYWJsZVN0ZXBzIDo9IGN0eC5TdHJpbmdTbGljZSgiZGlzYWJsZSIpCgkJY29uZmlnRmlsZXMgOj0gY3R4LlN0cmluZ1NsaWNlKCJjb25maWciKQoKCQlkZWZhdWx0Q29uZiA6PSBjb25mLkdldENvbmZpZygiZGVmYXVsdC1jb25maWciKQoKCQltYXBDb25maWdzIDo9IG1hcFtzdHJpbmddY29uZmlnLkNvbmZpZ3VyYXRpb257CgkJCSJkZWZhdWx0LWNvbmZpZyI6IGRlZmF1bHRDb25mLAoJCX0KCgkJZm9yIF8sIGNvbmZpZ0FyZyA6PSByYW5nZSBjb25maWdGaWxlcyB7CgkJCXYgOj0gc3RyaW5ncy5TcGxpdE4oY29uZmlnQXJnLCAiOiIsIDIpCgkJCWlmIGxlbih2KSA9PSAxIHsKCQkJCWFyZ3NDb25maWcgOj0gY29uZmlnLk5ld0NvbmZpZyhjb25maWcuQ29uZmlnRmlsZSh2WzBdKSkKCQkJCW1hcENvbmZpZ3NbImRlZmF1bHQtY29uZmlnIl0gPSBhcmdzQ29uZmlnCgkJCX0gZWxzZSBpZiBsZW4odikgPT0gMiB7CgkJCQlhcmdzQ29uZmlnIDo9IGNvbmZpZy5OZXdDb25maWcoY29uZmlnLkNvbmZpZ0ZpbGUodlsxXSkpCgkJCQltYXBDb25maWdzW3ZbMF1dID0gYXJnc0NvbmZpZwoJCQl9CgkJfQoKCQlkZWZhdWx0Q29uZiA9IG1hcENvbmZpZ3NbImRlZmF1bHQtY29uZmlnIl0KCgkJbG9nZ2VyQ29uZiA6PSBkZWZhdWx0Q29uZi5HZXRDb25maWcoImxvZ2dlciIpCgoJCWxvZ3J1c19tYXRlLkhpamFjaygKCQkJbG9ncnVzLlN0YW5kYXJkTG9nZ2VyKCksCgkJCWxvZ3J1c19tYXRlLldpdGhDb25maWcobG9nZ2VyQ29uZiksCgkJKQoKCQlmbG93Q3R4IDo9IGNvbnRleHQuTmV3Q29udGV4dCgpCgoJCWN0eExpc3QgOj0gY3R4LlN0cmluZ1NsaWNlKCJjdHgiKQoJCWN0eEZpbGVzIDo9IGN0eC5TdHJpbmdTbGljZSgiY3R4LWZpbGUiKQoKCQlsb2FkQ29udGV4dChjdHhMaXN0LCBjdHhGaWxlcywgZmxvd0N0eCkKCgkJaW5wdXRGaWxlcyA6PSBjdHguU3RyaW5nU2xpY2UoImlucHV0IikKCgkJZm9yIF8sIGlucHV0RmlsZSA6PSByYW5nZSBpbnB1dEZpbGVzIHsKCQkJdmFyIGlucHV0RmlsZURhdGEgW11ieXRlCgkJCWlucHV0RmlsZURhdGEsIGVyciA9IGlvdXRpbC5SZWFkRmlsZShpbnB1dEZpbGUpCgkJCWlmIGVyciAhPSBuaWwgewoJCQkJcmV0dXJuCgkJCX0KCgkJCXZhciBuYW1lVmFsdWVzIFtdZmxvdy5OYW1lVmFsdWUKCQkJZXJyID0ganNvbi5Vbm1hcnNoYWwoaW5wdXRGaWxlRGF0YSwgJm5hbWVWYWx1ZXMpCgkJCWlmIGVyciAhPSBuaWwgewoJCQkJcmV0dXJuCgkJCX0KCgkJCWZsb3cuQXBwZW5kT3V0cHV0KGZsb3dDdHgsIG5hbWVWYWx1ZXMuLi4pCgkJfQoKCQl0cmFucyA6PSBmbG93LkJlZ2luKGZsb3dDdHgsIGNvbmZpZy5XaXRoQ29uZmlnKGRlZmF1bHRDb25mKSkKCgkJZmxvd0xpc3QgOj0gY29uZi5HZXRTdHJpbmdMaXN0KCJmbG93IikKCgkJZmxvd0l0ZW1Db25maWcgOj0gY29uZi5HZXRDb25maWcoImNvbmZpZyIpCgoJCW1hcERpc2FiZWxTdGVwcyA6PSBtYXBbc3RyaW5nXWJvb2x7fQoKCQlmb3IgXywgc3RlcCA6PSByYW5nZSBkaXNhYmxlU3RlcHMgewoJCQltYXBEaXNhYmVsU3RlcHNbc3RlcF0gPSB0cnVlCgkJfQoKCQlmb3IgaSwgc3RyVVJMIDo9IHJhbmdlIGZsb3dMaXN0IHsKCgkJCXZhciBmbG93VVJMICp1cmwuVVJMCgkJCWZsb3dVUkwsIGVyciA9IHVybC5QYXJzZSgiZmxvdzovLyIgKyBzdHJVUkwpCgoJCQluYW1lIDo9IGZsb3dVUkwuSG9zdAoKCQkJaWQgOj0gZmxvd1VSTC5RdWVyeSgpLkdldCgiaWQiKQoKCQkJaWYgbGVuKGlkKSA9PSAwIHsKCQkJCWlkID0gc3RyY29udi5JdG9hKGkpCgkJCX0KCgkJCWlmIG1hcERpc2FiZWxTdGVwc1tpZF0gewoJCQkJY29udGludWUKCQkJfQoKCQkJaWYgaGFuZGxlckNvbmYsIGV4aXN0IDo9IG1hcENvbmZpZ3NbaWRdOyBleGlzdCB7CgkJCQl0cmFucy5UaGVuKG5hbWUsIGNvbmZpZy5XaXRoQ29uZmlnKGhhbmRsZXJDb25mKSkKCQkJfSBlbHNlIGlmIGZsb3dJdGVtQ29uZmlnLkhhc1BhdGgoaWQpIHsKCQkJCWhhbmRsZXJDb25mID0gZmxvd0l0ZW1Db25maWcuR2V0Q29uZmlnKGlkKQoJCQkJdHJhbnMuVGhlbihuYW1lLCBjb25maWcuV2l0aENvbmZpZyhoYW5kbGVyQ29uZikpCgkJCX0gZWxzZSB7CgkJCQl0cmFucy5UaGVuKG5hbWUpCgkJCX0KCQl9CgoJCWVyciA9IHRyYW5zLkNvbW1pdCgpCgoJCWlmIGVyciAhPSBuaWwgewoJCQlyZXR1cm4KCQl9CgoJCXF1aWV0IDo9IGN0eC5Cb29sKCJxdWlldCIpCgkJb3V0cHV0IDo9IGN0eC5TdHJpbmcoIm91dHB1dCIpCgoJCWlmICFxdWlldCB8fCBsZW4ob3V0cHV0KSA+IDAgewoKCQkJbmFtZVZhbHVlcyA6PSB0cmFucy5PdXRwdXQoKQoKCQkJdmFyIG91dGRhdGEgW11ieXRlCgkJCW91dGRhdGEsIGVyciA9IGpzb24uTWFyc2hhbEluZGVudCgKCQkJCW5hbWVWYWx1ZXMsCgkJCQkiIiwKCQkJCSIgICAgIikKCgkJCWlmIGVyciAhPSBuaWwgewoJCQkJcmV0dXJuCgkJCX0KCgkJCWlmIGxlbihvdXRwdXQpID09IDAgewoJCQkJaWYgbGVuKG5hbWVWYWx1ZXMpID4gMCB7CgkJCQkJZm10LlByaW50bG4oc3RyaW5nKG91dGRhdGEpKQoJCQkJfQoJCQl9IGVsc2UgewoJCQkJZXJyID0gaW91dGlsLldyaXRlRmlsZShvdXRwdXQsIG91dGRhdGEsIDA2NDQpCgkJCQlpZiBlcnIgIT0gbmlsIHsKCQkJCQlyZXR1cm4KCQkJCX0KCQkJfQoJCX0KCgkJcmV0dXJuCgl9Cn0KCmZ1bmMgbG9hZEVOVihjdHggKmNsaS5Db250ZXh0KSAoZXJyIGVycm9yKSB7CgllbnZzIDo9IGN0eC5TdHJpbmdTbGljZSgiZW52IikKCWVudkZpbGVzIDo9IGN0eC5TdHJpbmdTbGljZSgiZW52LWZpbGUiKQoKCWlmIGxlbihlbnZzKSA9PSAwICYmIGxlbihlbnZGaWxlcykgPT0gMCB7CgkJcmV0dXJuCgl9CgoJbWFwRU5WIDo9IG1hcFtzdHJpbmddc3RyaW5ne30KCglmb3IgXywgZW52IDo9IHJhbmdlIGVudnMgewoJCXYgOj0gc3RyaW5ncy5TcGxpdE4oZW52LCAiOiIsIDIpCgkJaWYgbGVuKHYpICE9IDIgewoJCQllcnIgPSBmbXQuRXJyb3JmKCJlbnYgZm9ybWF0IGVycm9yOiVzIiwgZW52KQoJCQlyZXR1cm4KCQl9CgoJCW1hcEVOVlt2WzBdXSA9IHZbMV0KCX0KCglmb3IgXywgZiA6PSByYW5nZSBlbnZGaWxlcyB7CgoJCXZhciBkYXRhIFtdYnl0ZQoJCWRhdGEsIGVyciA9IGlvdXRpbC5SZWFkRmlsZShmKQoJCWlmIGVyciAhPSBuaWwgewoJCQlyZXR1cm4KCQl9CgoJCWJ1ZiA6PSBieXRlcy5OZXdCdWZmZXIoZGF0YSkKCQlkZWNvZGVyIDo9IGpzb24uTmV3RGVjb2RlcihidWYpCgkJZGVjb2Rlci5Vc2VOdW1iZXIoKQoKCQl0bXBNYXAgOj0gbWFwW3N0cmluZ11zdHJpbmd7fQoJCWVyciA9IGRlY29kZXIuRGVjb2RlKCZ0bXBNYXApCgkJaWYgZXJyICE9IG5pbCB7CgkJCXJldHVybgoJCX0KCgkJZm9yIGssIHYgOj0gcmFuZ2UgdG1wTWFwIHsKCQkJbWFwRU5WW2tdID0gdgoJCX0KCX0KCglmb3IgaywgdiA6PSByYW5nZSBtYXBFTlYgewoJCW9zLlNldGVudihrLCB2KQoJfQoKCXJldHVybgp9CgpmdW5jIGxvYWRDb250ZXh0KGN0eExpc3QgW11zdHJpbmcsIGN0eEZpbGVzIFtdc3RyaW5nLCBmbG93Q3R4IGNvbnRleHQuQ29udGV4dCkgKGVyciBlcnJvcikgewoKCWlmIGxlbihjdHhMaXN0KSA9PSAwICYmIGxlbihjdHhGaWxlcykgPT0gMCB7CgkJcmV0dXJuCgl9CgoJbWFwQ3R4IDo9IG1hcFtzdHJpbmddc3RyaW5ne30KCglmb3IgXywgYyA6PSByYW5nZSBjdHhMaXN0IHsKCQl2IDo9IHN0cmluZ3MuU3BsaXROKGMsICI6IiwgMikKCQlpZiBsZW4odikgIT0gMiB7CgkJCWVyciA9IGZtdC5FcnJvcmYoImN0eCBmb3JtYXQgZXJyb3I6JXMiLCBjKQoJCQlyZXR1cm4KCQl9CgoJCW1hcEN0eFt2WzBdXSA9IHZbMV0KCX0KCglmb3IgXywgZiA6PSByYW5nZSBjdHhGaWxlcyB7CgoJCXZhciBkYXRhIFtdYnl0ZQoJCWRhdGEsIGVyciA9IGlvdXRpbC5SZWFkRmlsZShmKQoJCWlmIGVyciAhPSBuaWwgewoJCQlyZXR1cm4KCQl9CgoJCWJ1ZiA6PSBieXRlcy5OZXdCdWZmZXIoZGF0YSkKCQlkZWNvZGVyIDo9IGpzb24uTmV3RGVjb2RlcihidWYpCgkJZGVjb2Rlci5Vc2VOdW1iZXIoKQoKCQl0bXBNYXAgOj0gbWFwW3N0cmluZ11zdHJpbmd7fQoJCWVyciA9IGRlY29kZXIuRGVjb2RlKCZ0bXBNYXApCgkJaWYgZXJyICE9IG5pbCB7CgkJCXJldHVybgoJCX0KCgkJZm9yIGssIHYgOj0gcmFuZ2UgdG1wTWFwIHsKCQkJbWFwQ3R4W2tdID0gdgoJCX0KCX0KCglmb3IgaywgdiA6PSByYW5nZSBtYXBDdHggewoJCWZsb3dDdHguV2l0aFZhbHVlKGssIHYpCgl9CgoJcmV0dXJuCn0KCmZ1bmMgZ2VuZXJhdGVDb21tYW5kcyhjbWRzICpbXWNsaS5Db21tYW5kLCBuYW1lIHN0cmluZywgY29uZiBjb25maWcuQ29uZmlndXJhdGlvbikgewoKCWtleXMgOj0gY29uZi5LZXlzKCkKCglpZiBsZW4oa2V5cykgPT0gMCB7CgkJcmV0dXJuCgl9CgoJb2JqQ291bnQgOj0gMAoJZm9yIF8sIGtleSA6PSByYW5nZSBrZXlzIHsKCQlpZiBjb25mLklzT2JqZWN0KGtleSkgfHwga2V5ID09ICJ1c2FnZSIgewoJCQlvYmpDb3VudCsrCgkJfQoJfQoKCS8vIENvbW1hbmQKCWlmIG9iakNvdW50ICE9IGxlbihrZXlzKSB7CgkJKmNtZHMgPSBhcHBlbmQoKmNtZHMsCgkJCWNsaS5Db21tYW5kewoJCQkJTmFtZTogIG5hbWUsCgkJCQlVc2FnZTogY29uZi5HZXRTdHJpbmcoInVzYWdlIiksCgkJCQlGbGFnczogW11jbGkuRmxhZ3sKCQkJCQljbGkuU3RyaW5nU2xpY2VGbGFnewoJCQkJCQlOYW1lOiAgImRpc2FibGUsIGQiLAoJCQkJCQlVc2FnZTogImRpc2FibGUgc3RlcHMsIGUuZy46IC1kIGRldm9wcy5hbGl5dW4uY3MuY2x1c3Rlci5kZWxldGVkLndhaXQgLWQgZGV2b3BzLmFsaXl1bi5jcy5jbHVzdGVyLnJ1bm5pbmcud2FpdCIsCgkJCQkJfSwKCQkJCQljbGkuU3RyaW5nU2xpY2VGbGFnewoJCQkJCQlOYW1lOiAgImNvbmZpZywgYyIsCgkJCQkJCVVzYWdlOiAibWFwcGluZyBjb25maWcgdG8gZmxvdyBzdGVwcywgZS5nLjogLS1jb25maWcgZGVmYXVsdC5jb25mIC0tY29uZmlnIDA6c3RlcDAuY29uZiIsCgkJCQkJfSwKCQkJCQljbGkuU3RyaW5nU2xpY2VGbGFnewoJCQkJCQlOYW1lOiAgImVudiIsCgkJCQkJCVVzYWdlOiAiZS5nLjogLS1lbnYgVVNFUjp0ZXN0IC0tZW52IFBXRDphc2RmIiwKCQkJCQl9LAoJCQkJCWNsaS5TdHJpbmdTbGljZUZsYWd7CgkJCQkJCU5hbWU6ICAiY3R4IiwKCQkJCQkJVXNhZ2U6ICJlLmcuOiAtLWN0eCBjb2RlOmdvZ2FwIC0tZW52IGhlbGxvOndvcmxkIiwKCQkJCQl9LAoJCQkJCWNsaS5TdHJpbmdTbGljZUZsYWd7CgkJCQkJCU5hbWU6ICAiZW52LWZpbGUiLAoJCQkJCQlVc2FnZTogImUuZy46IC0tZW52LWZpbGUgYS5qc29uIC0tZW52LWZpbGUgYi5qc29uIiwKCQkJCQl9LAoJCQkJCWNsaS5TdHJpbmdTbGljZUZsYWd7CgkJCQkJCU5hbWU6ICAiaW5wdXQsIGkiLAoJCQkJCQlVc2FnZTogImlucHV0IGZpbGUgZnJvbSBvdGhlcidzIG91dHB1dCBmb3IgaW5pdCB0aGlzIGZsb3csIGUuZy46IC1pIG91dHB1dDEuanNvbiAtaSBvdXRwdXQyLmpzb24iLAoJCQkJCX0sCgkJCQkJY2xpLlN0cmluZ0ZsYWd7CgkJCQkJCU5hbWU6ICAib3V0cHV0LCBvIiwKCQkJCQkJVXNhZ2U6ICJmaWxlbmFtZSBvZiBvdXRwdXQiLAoJCQkJCX0sCgkJCQkJY2xpLkJvb2xGbGFnewoJCQkJCQlOYW1lOiAgInF1aWV0LCBxIiwKCQkJCQkJVXNhZ2U6ICJiZSBxdWlldCwgbm8gb3V0cHV0IGRhdGEgcHJpbnQiLAoJCQkJCX0sCgkJCQl9LAoJCQkJQWN0aW9uOiBuZXdBY3Rpb24obmFtZSwgY29uZiksCgkJCX0sCgkJKQoKCQlyZXR1cm4KCX0KCgl2YXIgc3ViQ29tbWFuZHMgW11jbGkuQ29tbWFuZAoKCWZvciBfLCBrZXkgOj0gcmFuZ2UgY29uZi5LZXlzKCkgewoKCQlpZiBrZXkgPT0gInVzYWdlIiB7CgkJCWNvbnRpbnVlCgkJfQoKCQlnZW5lcmF0ZUNvbW1hbmRzKCZzdWJDb21tYW5kcywga2V5LCBjb25mLkdldENvbmZpZyhrZXkpKQoKCX0KCgljdXJyZW50Q29tbWFuZCA6PSBjbGkuQ29tbWFuZHsKCQlOYW1lOiAgICAgICAgbmFtZSwKCQlVc2FnZTogICAgICAgY29uZi5HZXRTdHJpbmcoInVzYWdlIiksCgkJU3ViY29tbWFuZHM6IHN1YkNvbW1hbmRzLAoJfQoKCSpjbWRzID0gYXBwZW5kKCpjbWRzLCBjdXJyZW50Q29tbWFuZCkKfQo="
)

func main() {
	app := cli.NewApp()

	app.Name = "go-flow"
	app.HelpName = "go-flow"
	app.HideVersion = true

	app.Commands = cli.Commands{
		cli.Command{
			Name:  "build",
			Usage: "build your own flow into binary",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Usage: "flow config file",
				},
			},
			Action: build,
		},

		cli.Command{
			Name:  "run",
			Usage: "run flow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Usage: "flow config file",
				},
			},
			Action:          run,
			SkipFlagParsing: true,
			SkipArgReorder:  true,
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "be verbose",
		},
	}

	app.RunAndExitOnError()
}

func createBuilder(appName string, verboseOfGoBuild, verboseOfGoGet bool, conf config.Configuration) (bu *builder.Builder, err error) {

	argsConf := ""

	if verboseOfGoGet {
		argsConf += "go-get = [\"-v\"]\n"
	}

	if verboseOfGoBuild {
		argsConf += "go-build = [\"-v\"]\n"
	}

	buildConfStr := fmt.Sprintf(`%s {
packages = %s
build.args {
    %s
  }
}`, appName,
		toPkgList(conf.GetStringList("packages")),
		argsConf,
	)

	goTmpl, err := base64.StdEncoding.DecodeString(flowTempl)
	if err != nil {
		return
	}

	tmpl, err := template.New(appName).Parse(string(goTmpl))
	if err != nil {
		return
	}

	b, err := builder.NewBuilder(
		builder.ConfigString(buildConfStr),
		builder.Template(tmpl),
	)

	if err != nil {
		return
	}

	bu = b
	return
}

func toPkgList(list []string) string {

	var values []string

	for _, item := range list {
		values = append(values, fmt.Sprintf("\"%s\"", item))
	}

	return "[" + strings.Join(values, ",") + "]"
}

func build(ctx *cli.Context) (err error) {

	configFile := ctx.String("config")

	if len(configFile) == 0 {
		err = fmt.Errorf("please input config file")
		return
	}

	conf := config.NewConfig(config.ConfigFile(configFile))

	appName := conf.GetString("app.name", "app")

	verbose := ctx.Parent().Bool("verbose")

	b, err := createBuilder(appName, verbose, verbose, conf)
	if err != nil {
		return
	}

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	err = b.Build(map[string]interface{}{"config_str": fmt.Sprintf("`%s`", string(configData))}, appName)

	return
}

func run(ctx *cli.Context) (err error) {

	set := flag.NewFlagSet("run", 0)

	confArg := set.String("config", "", "flow config file")

	err = set.Parse(ctx.Args()[0:2])
	if err != nil {
		return
	}

	configFile := *confArg

	if len(configFile) == 0 {
		err = fmt.Errorf("please input config file")
		return
	}

	conf := config.NewConfig(config.ConfigFile(configFile))

	appName := conf.GetString("app.name", "app")

	verbose := ctx.Parent().Bool("verbose")

	b, err := createBuilder(appName, false, verbose, conf)
	if err != nil {
		return
	}

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	base64ConfStr := base64.StdEncoding.EncodeToString(configData)

	err = b.Run(map[string]interface{}{"config_str": fmt.Sprintf("`%s`", base64ConfStr)}, appName, ctx.Args()[2:])

	return
}
