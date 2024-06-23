; (use-modules (guix store) (guix derivations))

; (let (
;   (builder (add-text-to-store "builder.sh" "echo IFD HMMMM > $out\n" '())))

;   (derivation store "ifdhmmm" bash `("-e" builder))
; )

(use-modules (guix utils))
(use-modules (guix store))
(use-modules (guix derivations))
(use-modules (gnu packages bash))

(let* (
  (store (open-connection))
  (builder (add-text-to-store store "my-builder.sh" "echo hello world > $out\n" '()))
)
  (derivation store "foo"
              bash `("-e" ,builder)
              #:inputs `((,bash) (,builder))
              #:env-vars '(("HOME" . "/homeless"))))
