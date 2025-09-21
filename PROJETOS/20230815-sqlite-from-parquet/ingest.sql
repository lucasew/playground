.load ./libparquet.so

create virtual table source using parquet('./userdata1.parquet');

select * from source;
