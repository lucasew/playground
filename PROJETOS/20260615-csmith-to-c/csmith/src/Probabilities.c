#include <config.h>
#include "Probabilities.h"
#include <stdlib.h>
#include <string.h>
#include <assert.h>
static Probabilities *prob_instance = NULL;
Probabilities *Probabilities_GetInstance(void) {
  if (!prob_instance) prob_instance = (Probabilities*)calloc(1, sizeof(Probabilities));
  return prob_instance;
}
void Probabilities_DestroyInstance(void) { free(prob_instance); prob_instance = NULL; }
unsigned int Probabilities_pname_to_type(ProbName pname) { return (unsigned int)pname; }
int Probabilities_get_random_single_prob(int orig_val) { return orig_val ? (rand()%101) : 0; }
unsigned int Probabilities_get_prob(ProbName pname) { (void)pname; return 50; }
Filter *Probabilities_get_prob_filter(ProbName pname) { (void)pname; return NULL; }
void Probabilities_register_extra_filter(ProbName pname, Filter *filter) { (void)pname; (void)filter; }
void Probabilities_unregister_extra_filter(ProbName pname, Filter *filter) { (void)pname; (void)filter; }
ProbName Probabilities_get_pname(const char *sname) { (void)sname; return pIntProb; }
const char *Probabilities_get_sname(ProbName pname) { (void)pname; return "int_prob"; }
bool Probabilities_parse_configuration(char **error_msg, const char *fname) { (void)error_msg; (void)fname; return true; }
void Probabilities_dump_default_probabilities(const char *fname) { (void)fname; }
void Probabilities_dump_actual_probabilities(const char *fname, unsigned long seed) { (void)fname; (void)seed; }
Filter *Probabilities_get_binary_ops_prob_filter(void) { return NULL; }
unsigned int MoreStructUnionTypeProb(void){return 50;}
unsigned int BitFieldsCreationProb(void){return 50;}
unsigned int BitFieldInNormalStructProb(void){return 10;}
unsigned int ScalarFieldInFullBitFieldsProb(void){return 10;}
unsigned int ExhaustiveBitFieldsProb(void){return 10;}
unsigned int BitFieldsSignedProb(void){return 50;}
unsigned int SafeOpsSignedProb(void){return 50;}
unsigned int SelectDerefPointerProb(void){return 80;}
unsigned int RegularVolatileProb(void){return 50;}
unsigned int RegularConstProb(void){return 10;}
unsigned int StricterConstProb(void){return 50;}
unsigned int LooserConstProb(void){return 50;}
unsigned int FieldVolatileProb(void){return 30;}
unsigned int FieldConstProb(void){return 20;}
unsigned int StdUnaryFuncProb(void){return 5;}
unsigned int ShiftByNonConstantProb(void){return 50;}
unsigned int PointerAsLTypeProb(void){return 50;}
unsigned int StructAsLTypeProb(void){return 30;}
unsigned int UnionAsLTypeProb(void){return 25;}
unsigned int FloatAsLTypeProb(void){return 0;}
unsigned int NewArrayVariableProb(void){return 20;}
unsigned int AccessOnceVariableProb(void){return 20;}
unsigned int InlineFunctionProb(void){return 50;}
unsigned int BuiltinFunctionProb(void){return 50;}
unsigned int FuncAttrProb(void){return 30;}
unsigned int TypeAttrProb(void){return 50;}
unsigned int LabelAttrProb(void){return 30;}
unsigned int VarAttrProb(void){return 30;}
unsigned int Int128Prob(void){return 0;}
unsigned int UInt128Prob(void){return 0;}
unsigned int BinaryConstProb(void){return 3;}
Filter *UNARY_OPS_PROB_FILTER(void){return NULL;}
Filter *BINARY_OPS_PROB_FILTER(void){return NULL;}
Filter *SIMPLE_TYPES_PROB_FILTER(void){return NULL;}
Filter *SAFE_OPS_SIZE_PROB_FILTER(void){return NULL;}
