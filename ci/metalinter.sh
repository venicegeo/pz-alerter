#!/bin/sh

gometalinter \
--deadline=45s \
--concurrency=8 \
--vendor \
./...

#--exclude="exported method|const|type|function [A-Za-z\.0-9]* should have comment" \
#--exclude="cyclomatic complexity 13 of function createQueryDslAsString" \
#--exclude="cyclomatic complexity 12 of function \(\*Service\)\.GetMessage" \
#--exclude="cyclomatic complexity 11 of function \(\*Client\)\.GetMessages" \


