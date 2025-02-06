generate:
	gopapageno -l query_xpath.l -g query_xpath.g -log
generate-types:
	gopapageno -l query_xpath.l -g query_xpath.g -log -o ../../ext/xpath -types-only
test:
	python test_driver.py
clean:
	rm -rf *.pg.go query_xpath
