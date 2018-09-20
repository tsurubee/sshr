FROM kinogmt/centos-ssh

RUN useradd tsurubee && \
    echo "tsurubee:testpass" | chpasswd && \
    mkdir -p /home/tsurubee/.ssh

ADD misc/testdata/sshr_keys/id_rsa.pub /home/tsurubee/.ssh

RUN touch /home/tsurubee/.ssh/authorized_keys && \
    chown tsurubee:tsurubee /home/tsurubee/.ssh/authorized_keys && \
    chmod 600 /home/tsurubee/.ssh/authorized_keys && \
    cat /home/tsurubee/.ssh/id_rsa.pub > /home/tsurubee/.ssh/authorized_keys