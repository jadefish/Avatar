FROM postgres:11.1-alpine

ENV POSTGRES_INITDB_ARGS --encoding=UNICODE --lc-collate=C --lc-ctype=C

RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.9/main" > /etc/apk/repositories \
	&& apk add --no-cache \
	    perl=5.26.3-r0 \
	    perl-utils=5.26.3-r0 \
	    postgresql-plperl=11.1-r0 \
	    postgresql-plperl-contrib=11.1-r0 \
	    make \
	&& rm -fr /var/cache/apk/* \
	&& mv /usr/lib/postgresql/* /usr/local/lib/postgresql/ \
    && mv /usr/share/postgresql/extension/* /usr/local/share/postgresql/extension/ \
    && cpan App::cpanminus \
    && cpanm --no-wget Email::Valid
