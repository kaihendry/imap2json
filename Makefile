html.go: html/index.m4.html html/index.html html/jquery.js html/main.js
	@echo -e "package main\n\nconst html = \`" > $@
	@m4 -PE html/index.m4.html >> $@
	@echo -e "\n\`" >> $@

clean:
	rm -f html.go
