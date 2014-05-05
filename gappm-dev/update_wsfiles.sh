#!/bin/bash
cd "$(dirname "$0")"
gobytes -var=V_index_html -pack=webfiles ./../webclient/index.html > ./../gappm/webfiles/index.html.go
gobytes -var=V_main_css -pack=webfiles ./../webclient/main.css > ./../gappm/webfiles/main.css.go
gobytes -var=V_main_js -pack=webfiles ./../webclient/main.js > ./../gappm/webfiles/main.js.go
gobytes -var=V_pshubcl_js -pack=webfiles ./../webclient/pshubcl.js > ./../gappm/webfiles/pshubcl.js.go
#
##
### https://github.com/h8liu/gobytes/blob/master/main.go
##
#