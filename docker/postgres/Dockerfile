FROM postgres:14-bookworm AS postgres

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
            libdbd-pg-perl \
            libtap-parser-sourcehandler-pgtap-perl \
            postgresql-14-pgtap \
            pgtap && \
            rm -rf /var/lib/apt/lists/*


COPY docker/postgres/create-pgtap-extension.sh /docker-entrypoint-initdb.d
COPY docker/postgres/entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh && ln -s /usr/local/bin/entrypoint.sh /

ENTRYPOINT ["entrypoint.sh"]
CMD ["postgres"]
