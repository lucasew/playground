#ifndef ATTRIBUTE_H
#define ATTRIBUTE_H

#include "StdLibAliases.h"

struct AttributeGenerator;

/* C port: flattened attribute types using kind tag + union-ish fields.
   No inheritance, use char* for names, manual vector for choices. */
typedef enum {
  eAttrBoolean = 0,
  eAttrMultiChoice,
  eAttrAligned,
  eAttrSection,
} eAttributeKind;

typedef struct Attribute {
  eAttributeKind kind;
  char *name;   /* owned */
  int prob;
  /* extra per kind */
  int alignment; /* for aligned */
  char **choices; /* for multichoice, null term */
  int num_choices;
} Attribute;

Attribute *Attribute_new(const char *name, int prob, eAttributeKind kind);
char *Attribute_make_random(Attribute *a); /* returns new string, caller free */
void Attribute_delete(Attribute *a);

/* specific ctors helpers */
Attribute *BooleanAttribute_new(const char *name, int prob);
Attribute *MultiChoiceAttribute_new(const char *name, int prob, const char **choices, int nchoices);
Attribute *AlignedAttribute_new(const char *name, int prob, int alignment);
Attribute *SectionAttribute_new(const char *name, int prob);

typedef struct AttributeGenerator {
  Attribute **attributes;
  int count;
  int capacity;
} AttributeGenerator;

void AttributeGenerator_init(AttributeGenerator *g);
void AttributeGenerator_Output(AttributeGenerator *g, FILE *out);
void AttributeGenerator_add(AttributeGenerator *g, Attribute *a);
void AttributeGenerator_clear(AttributeGenerator *g);

#endif
