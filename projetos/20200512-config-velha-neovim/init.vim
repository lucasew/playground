let username = getenv("user")
let isNote = username == "lucasew" " Estou rodando o script no meu note?
let useClangd = 1
" let g:loaded_python3_provider = 0

" Inicializar map dos langservers
let g:LanguageClient_serverCommands = {}

" Inicializar configurações do lightline
let g:lightline = {}
let g:lightline.active = {}
let g:lightline.active.left = [ ['mode', 'paste'], ['readonly', 'filename', 'modified'] ]
let g:lightline.component = {}


" Baixar vimplug automagicamente
function! PreparaPlug(path)
    if empty(glob(a:path))
        execute '!wget -O ' . a:path . '  https://raw.github.com/junegunn/vim-plug/master/plug.vim'
        echom "Abra novamente e chame :PlugInstall"
    endif
endfunction

" Plug: Instalação
if has("nvim")
    call PreparaPlug("~/.config/nvim/autoload/plug.vim")
else
    call PreparaPlug("~/.vim/autoload/plug.vim")
endif

function! IsAC() " Checa se o note tá na tomada
    return readfile('/sys/class/power_supply/ACAD/online')
endfunction

com! Dosify set ff=dos

" Clipboard:
map gy "+y
map gp "+p
map gd "+d

" Mover pelas tabs
map <Leader>n <esc>:tabprevious<CR>
map <Leader>m <esc>:tabnext<CR>

" Tirar highlight da última pesquisa
noremap <C-n> :nohl<CR>

" Recarregar: vimrc ao salvar 
if has('nvim')
    autocmd! bufwritepost init.vim source %
else
    autocmd! bufwritepost .vimrc source %
endif

set encoding=utf-8 " Sempre usar utf-8 ao salvar os arquivos
set nu " Linhas numeradas
set showmatch " Highlight de parenteses e chaves
set path+=** " Busca recursiva

" Indentação
set autoindent " Mantem os niveis de indentação
set tabstop=4 " Tab de 4 espaços
set softtabstop=4
set shiftwidth=4
set shiftround
set expandtab " Tabs viram espaços
set list " Ilustra a identação

set nobackup " Desativar backup

" Corretor: Ortográfico
if isNote
    function! FzfSpellSink(word)
        exe 'normal! "_ciw'.a:word
    endfunction
    function! FzfSpell()
        let suggestions = spellsuggest(expand("<cword>"))
        return fzf#run({'source': suggestions, 'sink': function("FzfSpellSink"), 'down': 10 })
    endfunction
    nnoremap <tab> :call FzfSpell()<CR>
    autocmd BufEnter *.md,*.tex,*.txt set spell spelllang=pt_br
endif

set nocompatible " Desativando retrocompatibilidade com o vi
set mouse=a " Ativar mouse
set completeopt=menuone,noinsert,noselect " Customizações no menu de autocomplete, :help completeopt para mais info
" janela de preview que mostra algumas coisas dos comandos
" set completeopt+=preview " Ativa
set previewheight=3 " Altura máxima do preview
set winfixheight " Mantém

" Wildmenu: autocomplete para modo de comando
set wildmenu
set wildmode=list:longest,full

" Wildmenu: ignorar quem?
set wildignore+=*.pyc " Python
set wildignore+=*.o " C
set wildignore+=*.class " Java

syntax on " Ativa syntax highlight
filetype plugin on " Plugins necessitam disso
tab ball " Deixa menos bagunçado colocando um arquivo por aba

call plug#begin()
" Menos dor de cabeça, recomendo.
Plug 'lucasew/nocapsquit.vim'

" Plug 'jiangmiao/auto-pairs' " Fecha os blocos que abre, fica parecido com o esquema do vs code
Plug 'tpope/vim-surround' " Mexe com coisas em volta, tipo parenteses
Plug 'tomtom/tcomment_vim' " Preguiça de comentar as coisas na mão: gc {des,}comenta o selecionado, gcc {des,}comenta a linha
Plug 'junegunn/fzf', {'do': './install --all' }

" Snippets:
Plug 'honza/vim-snippets'
Plug 'SirVer/ultisnips'
let g:UltiSnipsExpandTrigger="<tab>"
let g:UltiSnipsJumpForwardTrigger="<c-n>"
let g:UltiSnipsJumpBackwardTrigger="<c-p>"

if has('nvim')
    Plug 'ncm2/ncm2' " Autocomplete
    Plug 'ncm2/ncm2-path' " Completa pastas e arquivos
    Plug 'roxma/nvim-yarp' " Dependencia do plugin anterior
    Plug 'ncm2/ncm2-syntax' " Completa pela definição de sintaxe
    Plug 'shougo/neco-syntax' " Dependencia
endif

if has('nvim')
    Plug 'autozimu/LanguageClient-neovim', {
                \ 'branch': 'next',
                \ 'do': 'bash install.sh',
                \ }
endif

" Vue:
Plug 'posva/vim-vue' " Syntax highlight para vue já puxando tudo certin

" Toml:
Plug 'cespare/vim-toml' " Syntax highlight toml

" Typescript:
if isNote
    Plug 'HerringtonDarkholme/yats.vim' " Syntax typescript
    " Plug 'ncm2/nvim-typescript', {'do': './install.sh'}
    " Plug 'mhartington/nvim-typescript', {'do': './install.sh'}
endif

" CtrlP
if isNote
    Plug 'ctrlpvim/ctrlp.vim'
endif

" Denite:
if 0
    Plug 'Shougo/denite.nvim'
    autocmd VimEnter call denite#custom#option('default', 'split', 'floating') " Janela flutuante do neovim
    autocmd VimEnter call denite#custom#option('default', 'prompt', 'λ:') " Prompt diferenciado

    autocmd FileType denite call s:denite_conf()
    function! s:denite_conf() abort
        nnoremap <silent><buffer><expr> <CR>
                    \ denite#do_map('do_action')
        nnoremap <silent><buffer><expr> d
                    \ denite#do_map('do_action', 'delete')
        nnoremap <silent><buffer><expr> p
                    \ denite#do_map('do_action', 'preview')
        nnoremap <silent><buffer><expr> q
                    \ denite#do_map('quit')
        nnoremap <silent><buffer><expr> i
                    \ denite#do_map('open_filter_buffer')
        nnoremap <silent><buffer><expr> <Space>
                    \ denite#do_map('toggle_select').'j'
    endfunction
endif


" VimScript:
Plug 'ncm2/ncm2-vim'
Plug 'Shougo/neco-vim'

" Markdown:
if executable('npm')
    " MarkdownPreview chama o preview
    Plug 'iamcco/markdown-preview.nvim', { 'do': 'cd app & npm install'  }
endif

" Latex:
if executable('pdflatex')
    " LLPStartPreview mostra o preview
    Plug 'xuhdev/vim-latex-live-preview', { 'for': 'tex' }
    Plug 'lervag/vimtex'
    " https://github.com/lervag/vimtex/issues/1160
    au User Ncm2Plugin call ncm2#register_source({
                \ 'name' : 'vimtex',
                \ 'priority': 9,
                \ 'subscope_enable': 1,
                \ 'complete_length': 1,
                \ 'scope': ['tex'],
                \ 'mark': 'tex',
                \ 'word_pattern': '\w+',
                \ 'complete_pattern': g:vimtex#re#ncm,
                \ 'on_complete': ['ncm2#on_complete#omni', 'vimtex#complete#omnifunc'],
                \ })
endif

" if executable('typescript-language-server')
"     let g:LanguageClient_serverCommands.typescript = [exepath('typescript-language-server'), '--stdio']
" endif
if useClangd
    if executable('clangd')
        let g:LanguageClient_serverCommands.c = [exepath('clangd')]
        let g:LanguageClient_serverCommands.cpp = [exepath('clangd')]
    endif
else
    if executable('ccls')
        let g:LanguageClient_serverCommands.c = [exepath('ccls')]
        let g:LanguageClient_serverCommands.cpp = [exepath('ccls')]
    endif
endif

if executable('pyls')
    let g:LanguageClient_serverCommands.python = [exepath('pyls')]
endif

if executable('gopls')
    let g:LanguageClient_serverCommands.go = [exepath('gopls')]
endif

if executable('jdtls')
    let g:LanguageClient_serverCommands.java = [exepath('jdtls'), '-data', getcwd()]
endif

if executable('lua-lsp')
    let g:LanguageClient_serverCommands.lua = [exepath('lua-lsp')]
endif

if executable('rls')
    let g:LanguageClient_serverCommands.rust = [exepath('rls')]
endif

if executable('javascript-typescript-stdio')
    let g:LanguageClient_serverCommands.javascript = [exepath('javascript-typescript-stdio')]
endif

if has('nvim') && executable('npm')
    Plug 'ncm2/ncm2-tern',  {'do': 'npm install'}
endif

" Dart:
if executable('dart_language_server')
    Plug 'dart-lang/dart-vim-plugin' " Syntax Highlight dart
    let g:LanguageClient_serverCommands.dart = [exepath('dart_language_server')]
    let dart_html_in_string=v:true
    let dart_style_guide = 2
    let dart_format_on_save = 1
endif

" Java:
if has('nvim') && 0
    Plug 'ObserverOfTime/ncm2-jc2', {'for': ['java', 'jsp']}
    Plug 'artur-shaik/vim-javacomplete2', {'for': ['java', 'jsp']}
endif


" Arduino:
if executable('arduino')
    function! ArduinoStatusLine()
        let port = arduino#GetPort()
        let line = '[' . g:arduino_board . ']'
        if !empty(port)
            line = line . ' (' . port . ':' . g:arduino_serial_baud . ')'
        endif
        return line
    endfunction
    Plug 'stevearc/vim-arduino'
    let g:arduino_dir = '/usr/share/arduino'
    " let g:lightline.component.arduino = '%{ArduinoStatusLine()}'
    autocmd BufNewFile,BufRead *.ino call add(g:lightline.active.left[1], 'arduino')
endif

if executable("gofmt")
    " autocmd BufWrite *.go :%!gofmt " Passa gofmt automagicamente
endif

if executable("nasm")
    autocmd BufNewFile,BufRead *.asm set filetype=nasm
endif

if executable('pdftotext')
    " Ler pdf no vim
    :command! -complete=file -nargs=1 Rpdf :r !pdftotext -nopgbrk <q-args> -
endif

if has('nvim')
    " NCM2:
    autocmd BufEnter * call ncm2#enable_for_buffer() " Ativa pra galera
    nnoremap <F5> :call LanguageClient_contextMenu()<CR>
endif

Plug 'itchyny/lightline.vim'
let g:lightline.colorscheme = 'wombat'

" Colorscheme:
Plug 'joshdick/onedark.vim' " Onedark <3
autocmd VimEnter * colorscheme onedark

" Trocar variante do tema
com! Transparent hi Normal ctermbg=none
com! White hi Normal ctermbg=white
com! Black hi Normal ctermbg=black

" Leader == espaço
let mapleader = ','

" Emmet: macro para html
Plug 'mattn/emmet-vim'
autocmd VimEnter * EmmetInstall
let g:user_emmet_mode='a'

" Neomake:
Plug 'neomake/neomake'
autocmd VimEnter * call neomake#configure#automake('w')

" Syntastic:
if has('nvim')
    Plug 'vim-syntastic/syntastic'
    autocmd VimEnter set g:airline_section_y += %#warningmsg#
endif

" Echodoc: 
if has('nvim')
    Plug 'Shougo/echodoc.vim'
    let g:echodoc#enable_at_startup=1
    " set cmdheight=2
    set noshowmode
    let g:echodoc#type = "virtual"
endif

" Startify:
Plug 'mhinz/vim-startify'
if executable("fortune")
  let g:startify_custom_header =
              \ map(split(system('fortune brasil'), '\n'), '"   ". v:val')
endif

" Man:
Plug 'bruno-/vim-man'

" IndentLines:
Plug 'Yggdroot/indentLine'
let g:indentLine_setConceal = 0

" Polyglot: Pacotão de sintaxe e tudo mais
Plug 'sheerun/vim-polyglot'

" TerminalMode: 
tnoremap <Esc> <C-\><C-n>

call plug#end()
