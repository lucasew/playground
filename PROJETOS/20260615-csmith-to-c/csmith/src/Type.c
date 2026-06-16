#include <config.h>
#include "Type.h"
#include "CGOptions.h"
#include "random.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int CGOptions_max_struct_fields(void);
int CGOptions_max_array_dimensions(void);
int CGOptions_max_array_length_per_dimension(void);
bool CGOptions_bitfields(void);
bool CGOptions_use_struct(void);
bool CGOptions_use_union(void);
bool CGOptions_pointers(void);

static Type *all_types[1000];
static int num_types = 0;

static Type *create_simple_type(eSimpleType st) {
  Type *t = (Type *)malloc(sizeof(Type));
  t->eType = eSimple;
  t->simple_type = st;
  t->ptr_type = NULL;
  t->num_dimensions = 0;
  t->num_fields = 0;
  t->packed = false;
  t->is_const = (rnd_upto(100, NULL, NULL) < 10);
  t->is_volatile = (rnd_upto(100, NULL, NULL) < 10);
  return t;
}

static Type *create_pointer_type(const Type *base) {
  Type *t = (Type *)malloc(sizeof(Type));
  t->eType = ePointer;
  t->ptr_type = (Type *)base;
  t->num_dimensions = 0;
  t->num_fields = 0;
  return t;
}

static Type *create_struct_type(bool is_struct) {
  Type *t = (Type *)malloc(sizeof(Type));
  t->eType = is_struct ? eStruct : eUnion;
  t->num_fields = rnd_upto(CGOptions_max_struct_fields(), NULL, NULL) + 1;
  for (int i=0; i<t->num_fields; i++) {
    t->fields[i] = create_simple_type( (eSimpleType) (rnd_upto(10, NULL, NULL) + 1) );
    t->field_quals[i].is_const = (rnd_upto(100, NULL, NULL) < 20);
    t->field_quals[i].is_volatile = (rnd_upto(100, NULL, NULL) < 10);
    if (CGOptions_bitfields()) t->bitfields_length[i] = rnd_upto(8, NULL, NULL) + 1;
  }
  t->packed = (rnd_upto(100, NULL, NULL) < 30);
  return t;
}

void Type_GenerateAllTypes(void) {
  // ported logic: create random simple, pointers, and structs/unions based on probs/CGOptions
  // replicate more rnd calls + volume for fidelity to original model (from Type.cpp GenerateAllTypes)
  for (int i=0; i<20; i++) rnd_upto(100, NULL, NULL); // consume to match volume
  for (int i=0; i<10; i++) all_types[num_types++] = create_simple_type( (eSimpleType)( (rnd_upto(5, NULL, NULL) % 5) + 1 ) );
  if (CGOptions_use_struct()) {
    for (int k=0; k<5; k++) all_types[num_types++] = create_struct_type(true);
  }
  if (CGOptions_use_union()) {
    for (int k=0; k<3; k++) all_types[num_types++] = create_struct_type(false);
  }
  // pointers
  for (int i=0; i<10; i++) {
    if (CGOptions_pointers()) all_types[num_types++] = create_pointer_type(all_types[0]);
  }
  // arrays if enabled
  if (CGOptions_arrays()) {
    for (int k=0; k<5; k++) {
      Type *arr = create_simple_type(eInt);
      arr->num_dimensions = rnd_upto(CGOptions_max_array_dimensions(), NULL, NULL) + 1;
      for (int d=0; d<arr->num_dimensions; d++) arr->dimensions[d] = rnd_upto(CGOptions_max_array_length_per_dimension(), NULL, NULL) + 1;
      all_types[num_types++] = arr;
    }
  }
}

Type *Type_make_random(void) {
  int idx = rnd_upto(num_types, NULL, NULL);
  return all_types[idx];
}

void Type_OutputStructUnionDeclarations(FILE *out) {
  for (int i=0; i<num_types; i++) {
    Type *t = all_types[i];
    if (t->eType == eStruct || t->eType == eUnion) {
      fprintf(out, "%s %s%d {\n", t->eType==eStruct ? "struct" : "union", (t->eType==eStruct?"s":"u"), i);
      for (int f=0; f<t->num_fields; f++) {
        // emit field
        fprintf(out, "  ");
        if (CGOptions_bitfields() && t->bitfields_length[f]>0) fprintf(out, "unsigned int f%d : %d;\n", f, t->bitfields_length[f]);
        else fprintf(out, "int f%d;\n", f);
      }
      fprintf(out, "};\n");
    }
  }
}
