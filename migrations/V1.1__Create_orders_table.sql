create table orders (
                        symbol varchar(20),
                        id varchar(60),
                        accountid varchar(60),
                        price double precision,
                        status varchar(10),
                        side varchar(10),
                        time int8,
                        leverage bool,
                        quantity double precision
);
