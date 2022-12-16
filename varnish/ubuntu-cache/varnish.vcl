vcl 4.1;

import directors;

backend ub1 { .host = "mirror.uepg.br"; .port = "80"; }
backend ub2 { .host = "ubuntu.c3sl.ufpr.br"; .port = "80"; }
backend ub3 { .host = "mirror.ufam.edu.br"; .port = "80"; }
backend ub4 { .host = "mirror.ufscar.br"; .port = "80"; }
backend ub5 { .host = "sft.if.usp.br"; .port = "80"; }
backend ub6 { .host = "ubuntu-archive.locaweb.com.br"; .port = "80"; }
backend ub7 { .host = "ubuntu.itsbrasil.net"; .port = "80"; }
backend ub8 { .host = "ubuntu.letscloud.io"; .port = "80"; }

sub vcl_backend_response {
    # 5 minutos de cache por padrão
    set beresp.ttl = 5m;
    set beresp.grace = 1h;

    if (bereq.url ~ "\.deb$") {
        set beresp.ttl = 24h;
        set beresp.grace = 2d;
    }
    if (bereq.url ~ "Contents.*$" || bereq.url ~ "Packages.*$") {
        set beresp.ttl = 1h;
        set beresp.grace = 12h;
    }
}

sub vcl_init {
    new ubuntu = directors.round_robin();
    ubuntu.add_backend(ub1);
    ubuntu.add_backend(ub2);
    ubuntu.add_backend(ub3);
    ubuntu.add_backend(ub4);
    ubuntu.add_backend(ub5);
    ubuntu.add_backend(ub6);
    ubuntu.add_backend(ub7);
    ubuntu.add_backend(ub8);
}

sub vcl_recv {
    /* set req.backend_hint = ubuntu.backend(); */
    set req.backend_hint = ub1;
}
