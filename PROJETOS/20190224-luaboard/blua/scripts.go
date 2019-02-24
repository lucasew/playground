package blua

func GetScripts(lw *LuaWrapper) []LuaInitScript {
    scripts := []LuaInitScript{}
    scripts = append(scripts, mod_action(lw)...)
    scripts = append(scripts, mod_board(lw)...)
    scripts = append(scripts, mod_me(lw)...)
    scripts = append(scripts, mod_utils(lw)...)
    return scripts
}
