#include<stdio.h>
#include<stdlib.h>
#include<sys/stat.h>
#include<string.h>
#include<errno.h>
#include <unistd.h>

#ifndef BUF_SIZE
#define BUF_SIZE 1024
#endif

int main(int argc, char** argv) {
    for (int i = 0; i < argc; i++) {
        printf("argv: %s\n", argv[i]);
    }
    char* env_out = getenv("out");
    char* env_text = getenv("textPath");
    char* env_NIX_LOG_FD = getenv("NIX_LOG_FD");

    char pathBuf[256];
    sprintf(pathBuf, "/build/.attr-%s", env_text + 1);

    int nix_fd = atoi(env_NIX_LOG_FD);
    fprintf(stderr, "%s\n", env_text);
    fprintf(stderr, "%s\n", env_out);

    /* if (rename(env_text, env_out)) { */
    /*     fprintf(stderr, "cant rename: %s\n", strerror(errno)); */
    /*     exit(1); */
    /* } */
    /* exit(0); */

    FILE *fi = fopen(env_text, "rb");
    if (!fi) {
        fprintf(stderr, "cant open input file: %s\n", strerror(errno));
        exit(1);
    }
    FILE *fo = fopen(env_out, "wb");
    if (!fo) {
        fprintf(stderr, "cant open output file: %s\n", strerror(errno));
        exit(1);
    }
    char buf[BUF_SIZE];
    size_t sz;
    for(;;) {
        char c = getc(fi);
        if (feof(fi)) break;
        putc(c, fo);
        putc(c, stdout);
    }
    /* for (;;) { */
    /*     sz = fread(buf, BUF_SIZE, 1, fi); */
    /*     if (!sz) break; */
    /*     fwrite(buf, sz, 1, fo); */
    /* } */
    fclose(fi);
    fclose(fo);
    return 0;
    /* fprintf(stderr, "fd: %i\n", nix_fd); */

    /* FILE *f; */
    /* if ((f = fopen(env_out, "w")) == NULL) { */
    /*     fprintf(stderr, "can't open output file: %s\n", strerror(errno)); */
    /*     exit(1); */
    /* } */
    /* if (fprintf(f, "%s", env_text) == -1) { */
    /*     fprintf(stderr, "can't write to the output file: %s\n", strerror(errno)); */
    /*     exit(1); */
    /* } */
    /* fflush(f); */
    /* fclose(f); */
}
