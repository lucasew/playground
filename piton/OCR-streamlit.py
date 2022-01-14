#!/usr/bin/env nix-shell
#! nix-shell -i python -p streamlit python3Packages.opencv3 python3Packages.pillow python3Packages.numpy python3Packages.pytesseract

import streamlit as st
from streamlit import cli as stcli
import cv2
from PIL import Image
import os
import pytesseract
from pytesseract import Output
import numpy as np

path = os.getcwd()

def main():
    global path
    st.title("API - CPF Detect")
    activities = ["DetectText", "Sobre"]
    choice = st.sidebar.selectbox("Selecione atividade: ", activities)
    if choice == "DetectText":
        image_file = st.file_uploader("Imagem a ser analisada", type=['jpg', 'png'])
        if image_file is not None:
            our_image = Image.open(image_file)
            st.text("Imagem original")
            st.image(our_image)
        texto = st.text_input("Escreva o texto aqui")
        st.write(f"{texto}")

        task = ["DetectTexto"]
        feature_choice = st.sidebar.selectbox("Detectar o texto", task)
        if st.button("Processar"):
            if feature_choice == "DetectTexto":
                try:
                    new_img = np.array(our_image.convert('RGB'))
                    d = pytesseract.image_to_data(
                            new_img, 
                            output_type = Output.DICT,
                            lang="por")
                    n_boxes = len(d['level'])
                    overlay = new_img.copy()
                    for i in range(n_boxes):
                        text = d['text'][i]
                        if text == texto:
                            (x, y, w, h) = (
                                    d['left'][i], 
                                    d['top'][i],
                                    d['width'][i],
                                    d['height'][i])
                            (x1, y1, w1, h1) = (
                                    d['left'][i + 1], 
                                    d['top'][i + 1],
                                    d['width'][i + 1],
                                    d['height'][i + 1])
                            cv2.rectangle(overlay, (x, y), (x1 + w1, y1 + h1), (255, 0, 0), -1)
                    alpha = 0.4
                    img_new = cv2.addWeighted(overlay, alpha, new_img, 1 - alpha, 0)
                    r = 1000.0 / img_new.shape[1]
                    dim = (1000, int(img_new.shape[0] * r))
                    resized = cv2.resize(img_new, dim, interpolation = cv2.INTER_AREA)
                    st.image(resized, width=600)
                except Exception as e:
                    print(e)
                    st.info("Deu ruim")


    elif choice == "Sobre":
        st.subheader("Sobre API - CPF Detect")
        st.markdown("**Desenvolvido por**: lucasew")
        st.info("Contato: exemplo@email.com")

    pass

if __name__ == '__main__':
    import sys
    if st._is_running_with_streamlit:
        main()
    else:
        sys.argv = ["streamlit", "run", sys.argv[0]]
        sys.exit(stcli.main())
