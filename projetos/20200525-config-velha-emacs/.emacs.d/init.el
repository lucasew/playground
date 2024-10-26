;; comando para abrir este arquivo de dotfile
(defun dotfile ()
  "Opens the dotfile for editing"
  (interactive)
  (find-file "~/.emacs.d/init.el"))

(defun linux-p ()
  (eq system-type "gnu/linux"))

;; ido-mode
(require 'ido)
(setq ido-enable-flex-matching t)
(setq ido-everywhere t)
(ido-mode 1)

(setq inhibit-startup-message t)
(tool-bar-mode -1) ;; tira toolbar
(menu-bar-mode -1) ;; tira barra de menu
(scroll-bar-mode -1) ;; tira barra de scroll
(global-hl-line-mode t) ;; marca linha do cursor
(line-number-mode t) ;; mostra linhas
(show-paren-mode 1) ;; mostra o parenteses do outro lado
(global-linum-mode t) ;; set nu do emacs
(electric-pair-mode 1) ;; autoclose dos bgl
(setq make-backup-files nil) ;; desativa aqueles arquivos que comecam com ~
(setq auto-save-default nil) ;; desativa aqueles #arquivos#
(setq straight-repository-branch "master") ;; branch straight.el
;;(setq straight-check-for-modifications t) ;; jeito diferenciado do straight puxar as modificações só que pode ser mais lento
(setq straight-use-package-by-default t) ;; use-package usando straight
(when (linux-p)
    (setq straight-vc-git-default-protocol "ssh"))

;; bootstrap straight.el
(defvar bootstrap-version)
(let ((bootstrap-file
       (expand-file-name "straight/repos/straight.el/bootstrap.el" user-emacs-directory))
      (bootstrap-version 5))
  (unless (file-exists-p bootstrap-file)
    (with-current-buffer
	(url-retrieve-synchronously
	 "https://raw.githubusercontent.com/raxod502/straight.el/develop/install.el"
	 'silent 'inhibit-cookies)
      (goto-char (point-max))
      (eval-print-last-sexp)))
  (load bootstrap-file nil 'nomessage))

;; instala o use-package pelo straight.el
(straight-use-package 'use-package)

;; tema
(use-package atom-one-dark-theme
  :ensure t
  :config
  (load-theme 'atom-one-dark t))

;; undo-tree
(use-package undo-tree 
  :ensure t
  :config
  (global-undo-tree-mode))

(use-package company 
  :ensure t
  :config
  (add-hook 'after-init-hook 'global-company-mode))

(use-package company-lsp 
  :ensure t
  :after company
  :init
  (push 'company-lsp company-backends)
  (setq company-lsp-async t)
  )

(use-package lsp-mode
  :ensure t
  :config
  (add-hook 'after-init-hook #'lsp-deferred))

(use-package go-mode
  :ensure t
  :config
  (add-hook 'go-mode-hook 'lsp-deferred))

(use-package company-quickhelp 
  :ensure t
  :after company)

;; slime
(use-package slime 
  :ensure t
  :config
  (setq inferior-lisp-program "sbcl") ;; qual interpretador o slime vai chamar
  )

;; autocomplete slime
(use-package slime-company 
  :ensure t
  :after (slime company)
  :config
  (setq slime-company-completion 'fuzzy)
  (setq slime-company-after-completion 'slime-company-just-one-space))

;; python
(use-package elpy 
  :ensure t
  :after (company)
  :init
  (elpy-enable))

;; clojure
;; C-c C-k dá eval no buffer no REPL

(use-package cider 
  :ensure t
  :after (company)
  :init
  (add-hook 'cider-repl-mode-hook #'company-mode)
  (add-hook 'cider-mode-hook #'company-mode)
  (setq cider-show-error-buffer 'only-in-repl))

;; yasnippet
(use-package yasnippet 
  :ensure t
  :config
  (yas-global-mode 1))

(use-package yasnippet-snippets
  :ensure t
  :after yasnippet)

(use-package org
  :ensure t)

;; smex: M-x melhorado
(use-package smex 
  :ensure t
  :config
  (smex-initialize)
  (global-set-key (kbd "M-x") 'smex)
  (global-set-key (kbd "M-X") 'smex-major-mode-commands)
  ;; M-x antigo
  (global-set-key (kbd "C-c C-c M-x") 'execute-extended-command)
  )

(use-package company-emoji
  :ensure t
  :after company
  :config
  (add-to-list 'company-backends 'company-emoji))

(use-package emojify
  :ensure t
  :config
  (add-hook 'after-init-hook #'global-emojify-mode)
  (setq emojify-company-tooltips-p t))

(use-package all-the-icons
  :ensure t
  :init
  (setq inhibit-compacting-font-caches t)) ;; icones

(use-package all-the-icons-dired 
  :ensure t
  :after all-the-icons
  :config
  (add-hook 'dired-mode-hook 'all-the-icons-dired-mode)
  (add-hook 'dired-mode-hook 'dired-hide-details-mode))

(use-package dired-toggle
  :ensure t
  :bind (
	 ("<f3>" . #'dired-toggle)
	 :map dired-mode-map
	 ("q" . #'dired-toggle-quit)
	 ([remap dired-find-file] . #'dired-toggle-find-file)
	 ([remap dired-up-directory] . #'dired-toggle-up-directory)
	 ("C-c C-u" . #'dired-toggle-up-directory)
	 )
  :config
  (setq dired-toggle-window-size 32)
  (setq dired-toggle-window-side 'left))

(use-package dashboard 
  :ensure t
  :after all-the-icons
  :config
  (setq initial-buffer-choice (lambda () (get-buffer "*dashboard*")))
  (setq dashboard-startup-banner 'logo)
  (setq dashboard-center-content t)
  (setq dashboard-show-shortcuts nil)
  (setq dashboard-set-heading-icons t)
  (setq dashboard-set-file-icons t)
  (dashboard-setup-startup-hook))

(use-package fzf
  :ensure t)

(use-package magit
  :ensure t)
