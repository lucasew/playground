#ifndef PROBABILITIES_H
#define PROBABILITIES_H
#include "Filter.h"
#include <stdbool.h>

typedef enum ProbName {
  pMoreStructUnionProb,
  pBitFieldsCreationProb,
  pBitFieldsSignedProb,
  pBitFieldInNormalStructProb,
  pScalarFieldInFullBitFieldsProb,
  pExhaustiveBitFieldsProb,
  pSafeOpsSignedProb,
  pSelectDerefPointerProb,
  pRegularVolatileProb,
  pRegularConstProb,
  pStricterConstProb,
  pLooserConstProb,
  pFieldVolatileProb,
  pFieldConstProb,
  pStdUnaryFuncProb,
  pShiftByNonConstantProb,
  pPointerAsLTypeProb,
  pStructAsLTypeProb,
  pUnionAsLTypeProb,
  pFloatAsLTypeProb,
  pNewArrayVariableProb,
  pAccessOnceVariableProb,
  pInlineFunctionProb,
  pBuiltinFunctionProb,
  pArrayOOBProb,
  pStatementProb,
  pAssignProb,
  pBlockProb,
  pForProb,
  pIfElseProb,
  pInvokeProb,
  pReturnProb,
  pContinueProb,
  pBreakProb,
  pGotoProb,
  pArrayOpProb,
  pAssignOpsProb,
  pSimpleAssignProb,
  pMulAssignProb,
  pDivAssignProb,
  pRemAssignProb,
  pAddAssignProb,
  pSubAssignProb,
  pLShiftAssignProb,
  pRShiftAssignProb,
  pBitAndAssignProb,
  pBitXorAssignProb,
  pBitOrAssignProb,
  pPreIncrProb,
  pPreDecrProb,
  pPostIncrProb,
  pPostDecrProb,
  pUnaryOpsProb,
  pPlusProb,
  pMinusProb,
  pNotProb,
  pBitNotProb,
  pBinaryOpsProb,
  pAddProb,
  pSubProb,
  pMulProb,
  pDivProb,
  pModProb,
  pCmpGtProb,
  pCmpLtProb,
  pCmpGeProb,
  pCmpLeProb,
  pCmpEqProb,
  pCmpNeProb,
  pAndProb,
  pOrProb,
  pBitXorProb,
  pBitAndProb,
  pBitOrProb,
  pRShiftProb,
  pLShiftProb,
  pSimpleTypesProb,
  pVoidProb,
  pCharProb,
  pIntProb,
  pShortProb,
  pLongProb,
  pLongLongProb,
  pUCharProb,
  pUIntProb,
  pUShortProb,
  pULongProb,
  pULongLongProb,
  pFloatProb,
  pInt128Prob,
  pUInt128Prob,
  pSafeOpsSizeProb,
  pInt8Prob,
  pInt16Prob,
  pInt32Prob,
  pInt64Prob,
  pFuncAttrProb,
  pTypeAttrProb,
  pLabelAttrProb,
  pVarAttrProb,
  pBinaryConstProb,
} ProbName;

#define MAX_PROB_NAME (pStatementProb + 1)

unsigned int MoreStructUnionTypeProb(void);
unsigned int BitFieldsCreationProb(void);
unsigned int BitFieldInNormalStructProb(void);
unsigned int ScalarFieldInFullBitFieldsProb(void);
unsigned int ExhaustiveBitFieldsProb(void);
unsigned int BitFieldsSignedProb(void);
unsigned int SafeOpsSignedProb(void);
unsigned int SelectDerefPointerProb(void);
unsigned int RegularVolatileProb(void);
unsigned int RegularConstProb(void);
unsigned int StricterConstProb(void);
unsigned int LooserConstProb(void);
unsigned int FieldVolatileProb(void);
unsigned int FieldConstProb(void);
unsigned int StdUnaryFuncProb(void);
unsigned int ShiftByNonConstantProb(void);
unsigned int PointerAsLTypeProb(void);
unsigned int StructAsLTypeProb(void);
unsigned int UnionAsLTypeProb(void);
unsigned int FloatAsLTypeProb(void);
unsigned int NewArrayVariableProb(void);
unsigned int AccessOnceVariableProb(void);
unsigned int InlineFunctionProb(void);
unsigned int BuiltinFunctionProb(void);
unsigned int FuncAttrProb(void);
unsigned int TypeAttrProb(void);
unsigned int LabelAttrProb(void);
unsigned int VarAttrProb(void);
unsigned int Int128Prob(void);
unsigned int UInt128Prob(void);
unsigned int BinaryConstProb(void);

struct Filter *UNARY_OPS_PROB_FILTER(void);
struct Filter *BINARY_OPS_PROB_FILTER(void);
struct Filter *SIMPLE_TYPES_PROB_FILTER(void);
struct Filter *SAFE_OPS_SIZE_PROB_FILTER(void);

typedef struct Probabilities {
  int dummy[8];
} Probabilities;
Probabilities *Probabilities_GetInstance(void);
void Probabilities_DestroyInstance(void);
unsigned int Probabilities_pname_to_type(ProbName pname);
int Probabilities_get_random_single_prob(int orig_val);
unsigned int Probabilities_get_prob(ProbName pname);
struct Filter *Probabilities_get_prob_filter(ProbName pname);
void Probabilities_register_extra_filter(ProbName pname, Filter *filter);
void Probabilities_unregister_extra_filter(ProbName pname, Filter *filter);
ProbName Probabilities_get_pname(const char *sname);
const char *Probabilities_get_sname(ProbName pname);
bool Probabilities_parse_configuration(char **error_msg, const char *fname);
void Probabilities_dump_default_probabilities(const char *fname);
void Probabilities_dump_actual_probabilities(const char *fname, unsigned long seed);
struct Filter *Probabilities_get_binary_ops_prob_filter(void);

#endif
