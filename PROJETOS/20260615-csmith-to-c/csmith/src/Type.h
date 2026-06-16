#ifndef TYPE_H
#define TYPE_H
#include <stdio.h>
#include <stdbool.h>

typedef enum {
  eSimple = 0,
  ePointer,
  eArray,
  eStruct,
  eUnion
} eTypeDesc;

typedef enum {
  eVoid = 0,
  eChar,
  eShort,
  eInt,
  eLong,
  eLongLong,
  eUChar,
  eUShort,
  eUInt,
  eULong,
  eULongLong,
  eFloat,
  eDouble
} eSimpleType;

typedef struct CVQualifiers {
  bool is_const;
  bool is_volatile;
} CVQualifiers;

typedef struct Type {
  eTypeDesc eType;
  eSimpleType simple_type;
  struct Type *ptr_type;
  int num_dimensions;
  int dimensions[10];
  int num_fields;
  struct Type *fields[100];
  CVQualifiers field_quals[100];
  int bitfields_length[100];
  bool packed;
  bool is_const;
  bool is_volatile;
} Type;

void Type_GenerateAllTypes(void);
void Type_OutputStructUnionDeclarations(FILE *out);
Type *Type_make_random(void);
#endif
