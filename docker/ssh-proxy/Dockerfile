FROM golang:latest

RUN mkdir -p /home/tsurubee/.ssh /home/hoge/.ssh

ADD misc/testdata/sshr_keys/id_rsa /home/tsurubee/.ssh
ADD misc/testdata/sshr_keys/id_rsa /home/hoge/.ssh
ADD misc/testdata/client_keys/id_rsa.pub /home/tsurubee/.ssh
ADD misc/testdata/client_keys/id_rsa.pub /home/hoge/.ssh

RUN touch /home/tsurubee/.ssh/authorized_keys && \
    chmod 600 /home/tsurubee/.ssh/authorized_keys && \
    cat /home/tsurubee/.ssh/id_rsa.pub > /home/tsurubee/.ssh/authorized_keys

RUN touch /home/hoge/.ssh/authorized_keys && \
    chmod 600 /home/hoge/.ssh/authorized_keys && \
    cat /home/hoge/.ssh/id_rsa.pub > /home/hoge/.ssh/authorized_keys