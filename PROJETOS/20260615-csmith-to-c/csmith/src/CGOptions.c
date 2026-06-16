#include <config.h>
#include "CGOptions.h"
#include <string.h>
#include <stdlib.h>
#include <stdbool.h>
static char *my_strdup(const char *s){if(!s)s="";size_t l=strlen(s)+1;char *p=malloc(l);if(p)memcpy(p,s,l);return p;}

/* state for options that support set (for CLI --no-xxx etc) */
static bool s_arrays = true;
static bool s_use_struct = true;
static bool s_use_union = true;
static bool s_bitfields = true;
static int s_max_funcs = 10;

bool CGOptions_prefix_name(void){return false;}
bool CGOptions_random_based(void){return true;}
bool CGOptions_arrays(void){return s_arrays;}
bool CGOptions_jumps(void){return true;}
bool CGOptions_pointers(void){return true;}
bool CGOptions_volatiles(void){return true;}
bool CGOptions_consts(void){return true;}
bool CGOptions_global_variables(void){return true;}
bool CGOptions_int8(void){return true;}
bool CGOptions_uint8(void){return true;}
bool CGOptions_longlong(void){return true;}
bool CGOptions_math64(void){return true;}
bool CGOptions_enable_float(void){return false;}
bool CGOptions_ccomp(void){return false;}
bool CGOptions_allow_int64(void){return true;}
int CGOptions_max_funcs(void){return s_max_funcs;}
int CGOptions_max_params(void){return 5;}
int CGOptions_max_block_size(void){return 4;}
int CGOptions_max_blk_depth(void){return 5;}
int CGOptions_max_expr_depth(void){return 10;}
int CGOptions_max_struct_fields(void){return 10;}
int CGOptions_max_union_fields(void){return 5;}
int CGOptions_max_nested_struct_level(void){return 3;}
int CGOptions_max_indirect_level(void){return 5;}
int CGOptions_max_array_dimensions(void){return 3;}
int CGOptions_max_array_length_per_dimension(void){return 10;}
int CGOptions_max_array_length(void){return 256;}
int CGOptions_max_array_num_in_loop(void){return 4;}
int CGOptions_func1_max_params(void){return 3;}
int CGOptions_coverage_test_size(void){return 500;}
int CGOptions_max_exhaustive_depth(void){return -1;}
int CGOptions_max_split_files(void){return 0;}
int CGOptions_interested_facts(void){return 0;}
int CGOptions_stop_by_stmt(void){return -1;}
int CGOptions_inline_function_prob(void){return 50;}
int CGOptions_builtin_function_prob(void){return 50;}
int CGOptions_array_oob_prob(void){return 0;}
int CGOptions_null_pointer_dereference_prob(void){return 0;}
int CGOptions_dead_pointer_dereference_prob(void){return 0;}
bool CGOptions_compute_hash(void){return true;}
bool CGOptions_depth_protect(void){return false;}
bool CGOptions_wrap_volatiles(void){return false;}
bool CGOptions_allow_const_volatile(void){return true;}
bool CGOptions_avoid_signed_overflow(void){return true;}
bool CGOptions_fixed_struct_fields(void){return false;}
bool CGOptions_expand_struct(void){return false;}
bool CGOptions_use_struct(void){return s_use_struct;}
bool CGOptions_use_union(void){return s_use_union;}
bool CGOptions_compound_assignment(void){return true;}
bool CGOptions_paranoid(void){return false;}
bool CGOptions_quiet(void){return false;}
bool CGOptions_concise(void){return false;}
bool CGOptions_nomain(void){return false;}
bool CGOptions_dfs_exhaustive(void){return false;}
bool CGOptions_compact_output(void){return false;}
bool CGOptions_klee(void){return false;}
bool CGOptions_crest(void){return false;}
bool CGOptions_coverage_test(void){return false;}
bool CGOptions_packed_struct(void){return true;}
bool CGOptions_bitfields(void){return s_bitfields;}
bool CGOptions_sequence_name_prefix(void){return false;}
bool CGOptions_compatible_check(void){return false;}
bool CGOptions_math_notmp(void){return false;}
bool CGOptions_strict_float(void){return false;}
bool CGOptions_strict_const_arrays(void){return false;}
bool CGOptions_return_structs(void){return true;}
bool CGOptions_return_unions(void){return true;}
bool CGOptions_arg_structs(void){return true;}
bool CGOptions_arg_unions(void){return true;}
bool CGOptions_volatile_pointers(void){return true;}
bool CGOptions_const_pointers(void){return true;}
bool CGOptions_access_once(void){return false;}
bool CGOptions_strict_volatile_rule(void){return false;}
bool CGOptions_addr_taken_of_locals(void){return true;}
bool CGOptions_fresh_array_ctrl_var_names(void){return false;}
bool CGOptions_builtins(void){return false;}
bool CGOptions_dangling_global_ptrs(void){return true;}
bool CGOptions_divs(void){return true;}
bool CGOptions_muls(void){return true;}
bool CGOptions_accept_argc(void){return true;}
bool CGOptions_random_random(void){return false;}
bool CGOptions_step_hash_by_stmt(void){return false;}
bool CGOptions_blind_check_global(void){return false;}
bool CGOptions_const_as_condition(void){return false;}
bool CGOptions_match_exact_qualifiers(void){return false;}
bool CGOptions_no_return_dead_ptr(void){return true;}
bool CGOptions_hash_value_printf(void){return true;}
bool CGOptions_signed_char_index(void){return true;}
bool CGOptions_identify_wrappers(void){return false;}
bool CGOptions_mark_mutable_const(void){return false;}
bool CGOptions_force_globals_static(void){return true;}
bool CGOptions_force_non_uniform_array_init(void){return true;}
bool CGOptions_pre_incr_operator(void){return true;}
bool CGOptions_pre_decr_operator(void){return true;}
bool CGOptions_post_incr_operator(void){return true;}
bool CGOptions_post_decr_operator(void){return true;}
bool CGOptions_unary_plus_operator(void){return true;}
bool CGOptions_use_embedded_assigns(void){return true;}
bool CGOptions_use_comma_exprs(void){return true;}
bool CGOptions_take_union_field_addr(void){return true;}
bool CGOptions_vol_struct_union_fields(void){return true;}
bool CGOptions_const_struct_union_fields(void){return true;}
bool CGOptions_lang_cpp(void){return false;}
bool CGOptions_cpp11(void){return false;}
bool CGOptions_fast_execution(void){return false;}
bool CGOptions_func_attr_flag(void){return false;}
bool CGOptions_type_attr_flag(void){return false;}
bool CGOptions_label_attr_flag(void){return false;}
bool CGOptions_var_attr_flag(void){return false;}
bool CGOptions_Int128(void){return false;}
bool CGOptions_UInt128(void){return false;}
bool CGOptions_binary_constant(void){return false;}
const char *CGOptions_split_files_dir(void){return "./output";}
const char *CGOptions_output_file(void){return "";}
const char *CGOptions_struct_output(void){return "";}
const char *CGOptions_dfs_debug_sequence(void){return "";}
const char *CGOptions_partial_expand(void){return "";}
const char *CGOptions_delta_monitor(void){return "";}
const char *CGOptions_delta_output(void){return "";}
const char *CGOptions_go_delta(void){return "";}
const char *CGOptions_delta_input(void){return "";}
const char *CGOptions_dump_default_probabilities(void){return "";}
const char *CGOptions_dump_random_probabilities(void){return "";}
const char *CGOptions_probability_configuration(void){return "";}
const char *CGOptions_vol_tests_mach(void){return "";}
const char *CGOptions_conflict_msg(void){return "";}
int CGOptions_int_size(void){return sizeof(int);}
int CGOptions_pointer_size(void){return sizeof(void*);}
bool CGOptions_has_conflict(void){return false;}
bool CGOptions_has_random_based_conflict(void){return false;}
bool CGOptions_is_random(void){return true;}
bool CGOptions_has_extension_support(void){return false;}
bool CGOptions_x86_64(void){
#if defined __x86_64__
return true;
#else
return false;
#endif
}
bool CGOptions_set_vol_tests(const char *s){(void)s;return false;}
void CGOptions_monitored_funcs(const char *fnames){(void)fnames;}
void CGOptions_safe_math_wrapper(const char *ids){(void)ids;}
bool CGOptions_safe_math_wrapper_id(int id){(void)id;return true;}
void CGOptions_enable_builtin_kinds(const char *kinds){(void)kinds;}
void CGOptions_disable_builtin_kinds(const char *kinds){(void)kinds;}
bool CGOptions_enabled_builtin(const char *ks){(void)ks;return true;}
void CGOptions_fix_options_for_cpp(void){}
void CGOptions_set_platform_specific_options(void){}
void CGOptions_int_size_set(int p){(void)p;}
void CGOptions_pointer_size_set(int p){(void)p;}
const char *CGOptions_split_files_dir_set(const char *p){(void)p;return "";}
const char *CGOptions_output_file_set(const char *p){(void)p;return "";}
int CGOptions_max_funcs_set(int p){ s_max_funcs = p; return p; }
bool CGOptions_arrays_set(bool p){ s_arrays = p; return p; }
bool CGOptions_use_struct_set(bool p){ s_use_struct = p; return p; }
bool CGOptions_use_union_set(bool p){ s_use_union = p; return p; }
bool CGOptions_bitfields_set(bool p){ s_bitfields = p; return p; }
void CGOptions_set_default_settings(void){
  s_arrays = true;
  s_use_struct = true;
  s_use_union = true;
  s_bitfields = true;
  s_max_funcs = 10;
  /* other defaults as needed for full */
}
/* add more stubs as needed, return p or false */
bool CGOptions_compute_hash_set(bool p){return p;}
/* ... many more can be added but for build link, the used ones above suffice for stub run */
