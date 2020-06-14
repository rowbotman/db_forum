FROM golang:1.13-stretch AS lang

WORKDIR /home/db-park

COPY . .
RUN go get -d && go build -v

FROM ubuntu:18.04
ENV PGVERSION 10
RUN apt-get update && apt-get install -y postgresql-$PGVERSION

USER postgres
WORKDIR /home/db-park
RUN cd /home/db-park
COPY . .
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER park WITH SUPERUSER PASSWORD 'admin';" &&\
    createdb -O park park_forum && psql -d park_forum -c "CREATE EXTENSION IF NOT EXISTS citext;" &&\
    psql park_forum -a -f ./init.sql &&\
    /etc/init.d/postgresql stop

USER root

RUN echo "listen_addresses = '*'" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "synchronous_commit = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "fsync = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "full_page_writes = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "max_wal_size = 1GB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "shared_buffers = 512MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "effective_cache_size = 256MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "work_mem = 64MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "maintenance_work_mem = 128MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "unix_socket_directories = '/var/run/postgresql'" >> /etc/postgresql/$PGVERSION/main/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

EXPOSE 5000

USER postgres

WORKDIR /home/db-park
COPY --from=lang /home/db-park .

CMD /etc/init.d/postgresql start && ./db-park
