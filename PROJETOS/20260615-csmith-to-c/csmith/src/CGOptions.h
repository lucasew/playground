// -*- mode: C++ -*-
//
// Copyright (c) 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015, 2016, 2017 The
// University of Utah All rights reserved.
//
// This file is part of `csmith', a random generator of C programs.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//   * Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//
//   * Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer in the
//     documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

#ifndef CGOPTIONS_H
#define CGOPTIONS_H

#include "StdLibAliases.h"

///////////////////////////////////////////////////////////////////////////////

/*
* XXX --- Collect all the default values here.
*/
#define CGOPTIONS_DEFAULT_MAX_FUNCS 10
#define CGOPTIONS_DEFAULT_MAX_PARAMS 5
#define CGOPTIONS_DEFAULT_FUNC1_MAX_PARAMS 3
#define CGOPTIONS_DEFAULT_COVERAGE_TEST_SIZE 500
#define CGOPTIONS_DEFAULT_MAX_BLOCK_SIZE 4
#define CGOPTIONS_DEFAULT_MAX_BLOCK_DEPTH 5
#define CGOPTIONS_DEFAULT_MAX_EXPR_DEPTH 10
#define CGOPTIONS_DEFAULT_MAX_STRUCT_FIELDS 10
#define CGOPTIONS_DEFAULT_MAX_UNION_FIELDS 5
#define CGOPTIONS_DEFAULT_MAX_NESTED_STRUCT_LEVEL 3
#define CGOPTIONS_DEFAULT_MAX_INDIRECT_LEVEL 5
#define CGOPTIONS_DEFAULT_MAX_ARRAY_DIMENSIONS 3
#define CGOPTIONS_DEFAULT_MAX_ARRAY_LENGTH_PER_DIMENSION 10
#define CGOPTIONS_DEFAULT_MAX_ARRAY_LENGTH 256
#define CGOPTIONS_DEFAULT_MAX_ARRAY_NUM_IN_LOOP 4
#define CGOPTIONS_DEFAULT_MAX_EXHAUSTIVE_DEPTH (-1)
/* 0 means we output to the standard output */
#define CGOPTIONS_DEFAULT_MAX_SPLIT_FILES 0
#define CGOPTIONS_DEFAULT_SPLIT_FILES_DIR "./output"
#define CGOPTIONS_DEFAULT_OUTPUT_FILE ""
#define PLATFORM_CONFIG_FILE "platform.info"

/*
* C port: no class, functions defined in CGOptions.c with static state.
*/
bool CGOptions_compute_hash(void);
bool CGOptions_compute_hash_set(bool p);

bool CGOptions_depth_protect(void);
bool CGOptions_depth_protect_set(bool p);

int CGOptions_max_split_files(void);
int CGOptions_max_split_files_set(int p);

const char *CGOptions_split_files_dir(void);
const char *CGOptions_split_files_dir_set(const char *p);

const char *CGOptions_output_file(void);
const char *CGOptions_output_file_set(const char *p);

int CGOptions_max_funcs(void);
int CGOptions_max_funcs_set(int p);

int CGOptions_max_params(void);
int CGOptions_max_params_set(int p);

int CGOptions_max_block_size(void);
int CGOptions_max_block_size_set(int p);

int CGOptions_max_blk_depth(void);
int CGOptions_max_blk_depth_set(int p);

int CGOptions_max_expr_depth(void);
int CGOptions_max_expr_depth_set(int p);

bool CGOptions_wrap_volatiles(void);
bool CGOptions_wrap_volatiles_set(bool p);

bool CGOptions_allow_const_volatile(void);
bool CGOptions_allow_const_volatile_set(bool p);

bool CGOptions_avoid_signed_overflow(void);
bool CGOptions_avoid_signed_overflow_set(bool p);

int max_struct_fields();
int CGOptions_max_struct_fields_set(int p);

int max_union_fields();
int CGOptions_max_union_fields_set(int p);

int max_nested_struct_level();
int CGOptions_max_nested_struct_level_set(int p);

const char *struct_output();
const char *CGOptions_struct_output_set(const char *p);

bool fixed_struct_fields();
bool CGOptions_fixed_struct_fields_set(bool p);

bool expand_struct();
bool CGOptions_expand_struct_set(bool p);

bool use_struct();
bool CGOptions_use_struct_set(bool p);

bool use_union();
bool CGOptions_use_union_set(bool p);

int max_indirect_level();
int CGOptions_max_indirect_level_set(int p);

int max_array_dimensions();
int CGOptions_max_array_dimensions_set(int p);

int max_array_length_per_dimension();
int CGOptions_max_array_length_per_dimension_set(int p);

int max_array_length();
int CGOptions_max_array_length_set(int p);

bool CGOptions_compound_assignment(void);
bool CGOptions_compound_assignment_set(bool p);

int interested_facts();
int CGOptions_interested_facts_set(int p);

bool CGOptions_paranoid(void);
bool CGOptions_paranoid_set(bool p);

bool CGOptions_quiet(void);
bool CGOptions_quiet_set(bool p);

bool CGOptions_concise(void);
bool CGOptions_concise_set(bool p);

bool CGOptions_nomain(void);
bool CGOptions_nomain_set(bool p);

bool CGOptions_random_based(void);
bool CGOptions_random_based_set(bool p);

bool CGOptions_dfs_exhaustive(void);
bool CGOptions_dfs_exhaustive_set(bool p);

const char *CGOptions_dfs_debug_sequence(void);
const char *CGOptions_dfs_debug_sequence_set(const char *p);

int CGOptions_max_exhaustive_depth(void);
int CGOptions_max_exhaustive_depth_set(int p);

bool CGOptions_compact_output(void);
bool CGOptions_compact_output_set(bool p);

int CGOptions_func1_max_params(void);
int CGOptions_func1_max_params_set(int p);

bool CGOptions_klee(void);
bool CGOptions_klee_set(bool p);

bool CGOptions_crest(void);
bool CGOptions_crest_set(bool p);

bool CGOptions_ccomp(void);
bool CGOptions_ccomp_set(bool p);

bool CGOptions_coverage_test(void);
bool CGOptions_coverage_test_set(bool p);

int CGOptions_coverage_test_size(void);
int CGOptions_coverage_test_size_set(int p);

bool CGOptions_prefix_name(void);
bool CGOptions_prefix_name_set(bool p);

bool CGOptions_sequence_name_prefix(void);
bool CGOptions_sequence_name_prefix_set(bool p);

bool CGOptions_compatible_check(void);
bool CGOptions_compatible_check_set(bool p);

bool CGOptions_packed_struct(void);
bool CGOptions_packed_struct_set(bool p);

bool CGOptions_bitfields(void);
bool CGOptions_bitfields_set(bool p);

const char *CGOptions_partial_expand(void);
const char *CGOptions_partial_expand_set(const char *p);

const char *CGOptions_delta_monitor(void);
const char *CGOptions_delta_monitor_set(const char *p);

const char *CGOptions_delta_output(void);
const char *CGOptions_delta_output_set(const char *p);

const char *CGOptions_go_delta(void);
const char *CGOptions_go_delta_set(const char *p);

const char *CGOptions_delta_input(void);
const char *CGOptions_delta_input_set(const char *p);

bool CGOptions_no_delta_reduction(void);
bool CGOptions_no_delta_reduction_set(bool p);

bool CGOptions_math_notmp(void);
bool CGOptions_math_notmp_set(bool p);

bool CGOptions_math64(void);
bool CGOptions_math64_set(bool p);

bool CGOptions_inline_function(void);
bool CGOptions_inline_function_set(bool p);

bool CGOptions_longlong(void);
bool CGOptions_longlong_set(bool p);

bool CGOptions_int8(void);
bool CGOptions_int8_set(bool p);

bool CGOptions_uint8(void);
bool CGOptions_uint8_set(bool p);

bool CGOptions_enable_float(void);
bool CGOptions_enable_float_set(bool p);

bool CGOptions_strict_float(void);
bool CGOptions_strict_float_set(bool p);

bool CGOptions_pointers(void);
bool CGOptions_pointers_set(bool p);

bool CGOptions_arrays(void);
bool CGOptions_arrays_set(bool p);

bool CGOptions_strict_const_arrays(void);
bool CGOptions_strict_const_arrays_set(bool p);

bool CGOptions_jumps(void);
bool CGOptions_jumps_set(bool p);

bool CGOptions_return_structs(void);
bool CGOptions_return_structs_set(bool p);

bool CGOptions_return_unions(void);
bool CGOptions_return_unions_set(bool p);

bool CGOptions_arg_structs(void);
bool CGOptions_arg_structs_set(bool p);

bool CGOptions_arg_unions(void);
bool CGOptions_arg_unions_set(bool p);

bool CGOptions_volatiles(void);
bool CGOptions_volatiles_set(bool p);

bool CGOptions_volatile_pointers(void);
bool CGOptions_volatile_pointers_set(bool p);

bool CGOptions_const_pointers(void);
bool CGOptions_const_pointers_set(bool p);

bool CGOptions_global_variables(void);
bool CGOptions_global_variables_set(bool p);

const char *CGOptions_vol_tests_mach(void);
bool CGOptions_set_vol_tests(const char *s);

bool CGOptions_access_once_set(bool p);
bool CGOptions_access_once(void);

bool CGOptions_strict_volatile_rule_set(bool p);
bool CGOptions_strict_volatile_rule(void);

bool CGOptions_addr_taken_of_locals_set(bool p);
bool CGOptions_addr_taken_of_locals(void);

bool CGOptions_fresh_array_ctrl_var_names_set(bool p);
bool CGOptions_fresh_array_ctrl_var_names(void);

bool CGOptions_consts(void);
bool CGOptions_consts_set(bool p);

bool CGOptions_builtins(void);
bool CGOptions_builtins_set(bool p);

bool CGOptions_dangling_global_ptrs(void);
bool CGOptions_dangling_global_ptrs_set(bool p);

bool CGOptions_divs(void);
bool CGOptions_divs_set(bool p);

bool CGOptions_muls(void);
bool CGOptions_muls_set(bool p);

bool CGOptions_accept_argc(void);
bool CGOptions_accept_argc_set(bool p);

bool CGOptions_random_random(void);
bool CGOptions_random_random_set(bool p);

const char *CGOptions_dump_default_probabilities(void);
const char *CGOptions_dump_default_probabilities_set(const char *p);

const char *CGOptions_dump_random_probabilities(void);
const char *CGOptions_dump_random_probabilities_set(const char *p);

const char *CGOptions_probability_configuration(void);
const char *CGOptions_probability_configuration_set(const char *p);

bool CGOptions_step_hash_by_stmt(void);
bool CGOptions_step_hash_by_stmt_set(bool p);

bool CGOptions_blind_check_global(void);
bool CGOptions_blind_check_global_set(bool p);

int CGOptions_stop_by_stmt(void);
int CGOptions_stop_by_stmt_set(int p);

void monitored_funcs(const char *fnames);

bool CGOptions_const_as_condition(void);
bool CGOptions_const_as_condition_set(bool p);

bool CGOptions_no_return_dead_ptr(void);
bool CGOptions_no_return_dead_ptr_set(bool p);

bool CGOptions_hash_value_printf(void);
bool CGOptions_hash_value_printf_set(bool p);

bool CGOptions_signed_char_index(void);
bool CGOptions_signed_char_index_set(bool p);

/////////////////////////////////////////////////////////
void set_default_settings(void);

bool CGOptions_has_conflict(void);
bool CGOptions_has_random_based_conflict(void);
const char * *conflict_msg(void);

bool CGOptions_is_random(void);

bool has_extension_support();

bool allow_int64();

bool CGOptions_match_exact_qualifiers(void);
bool CGOptions_match_exact_qualifiers_set(bool p);

int max_array_num_in_loop();
int CGOptions_max_array_num_in_loop_set(int p);

bool x86_64();

bool CGOptions_identify_wrappers(void);
bool CGOptions_identify_wrappers_set(bool p);


bool CGOptions_mark_mutable_const(void);
bool CGOptions_mark_mutable_const_set(bool p);

bool CGOptions_force_globals_static(void);
bool CGOptions_force_globals_static_set(bool p);

bool CGOptions_force_non_uniform_array_init(void);
bool CGOptions_force_non_uniform_array_init_set(bool p);

int CGOptions_inline_function_prob(void);
int CGOptions_inline_function_prob_set(int p);

int CGOptions_builtin_function_prob(void);
int CGOptions_builtin_function_prob_set(int p);

int CGOptions_array_oob_prob(void);
int array_oob_prob(int);

int CGOptions_null_pointer_dereference_prob(void);
int CGOptions_null_pointer_dereference_prob_set(int p);

int CGOptions_dead_pointer_dereference_prob(void);
int CGOptions_dead_pointer_dereference_prob_set(int p);

bool CGOptions_pre_incr_operator(void);
bool CGOptions_pre_incr_operator_set(bool p);

bool CGOptions_pre_decr_operator(void);
bool CGOptions_pre_decr_operator_set(bool p);

bool CGOptions_post_incr_operator(void);
bool CGOptions_post_incr_operator_set(bool p);

bool CGOptions_post_decr_operator(void);
bool CGOptions_post_decr_operator_set(bool p);

bool CGOptions_unary_plus_operator(void);
bool CGOptions_unary_plus_operator_set(bool p);

bool CGOptions_use_embedded_assigns(void);
bool CGOptions_use_embedded_assigns_set(bool p);

bool CGOptions_use_comma_exprs(void);
bool CGOptions_use_comma_exprs_set(bool p);

bool CGOptions_take_union_field_addr(void);
bool CGOptions_take_union_field_addr_set(bool p);

bool CGOptions_vol_struct_union_fields(void);
bool CGOptions_vol_struct_union_fields_set(bool p);

bool CGOptions_const_struct_union_fields(void);
bool CGOptions_const_struct_union_fields_set(bool p);

int CGOptions_int_size(void);

int CGOptions_pointer_size(void);

void set_platform_specific_options(void);

bool CGOptions_lang_cpp(void);
bool CGOptions_lang_cpp_set(bool p);

bool CGOptions_cpp11(void);
bool CGOptions_cpp11_set(bool p);

void enable_builtin_kinds(const char *kinds);
void disable_builtin_kinds(const char *kinds);
bool enabled_builtin(const char *ks);

void fix_options_for_cpp(void);

bool CGOptions_fast_execution(void);
bool CGOptions_fast_execution_set(bool p);

// GCC C Extensions
bool CGOptions_func_attr_flag(void);
bool CGOptions_func_attr_flag_set(bool p);

bool CGOptions_type_attr_flag(void);
bool CGOptions_type_attr_flag_set(bool p);

bool CGOptions_label_attr_flag(void);
bool CGOptions_label_attr_flag_set(bool p);

bool CGOptions_var_attr_flag(void);
bool CGOptions_var_attr_flag_set(bool p);

bool CGOptions_Int128(void);
bool CGOptions_Int128_set(bool p);

bool CGOptions_UInt128(void);
bool CGOptions_UInt128_set(bool p);

bool CGOptions_binary_constant(void);
bool CGOptions_binary_constant_set(bool p);







// Until I do this right, just make them all static.


///////////////////////////////////////////////////////////////////////////////

#endif // CGOPTIONS_H

// Local Variables:
// c-basic-offset: 4
// tab-width: 4
// End:

// End of file.
