/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package tk.biglucas.ex1;

/**
 *
 * @author lucasew
 */
public class CD extends Item {
    private String artista;
    private Integer faixas;
    
    protected void middle_print() {
        System.out.printf("artista=%s ", artista);
        System.out.printf("faixas=%d ", faixas);
        super.middle_print();
    }
    
    public void print() {
        System.out.printf("CD(");
        this.middle_print();
        super.middle_print();
        System.out.println(")");
    }

    public CD() {
    }

    public CD(String artista, Integer faixas, String titulo, Double duracao, String comentarios) {
        super(titulo, duracao, comentarios);
        this.artista = artista;
        this.faixas = faixas;
    }

    public String getArtista() {
        return artista;
    }

    public Integer getFaixas() {
        return faixas;
    }

    public void setArtista(String artista) {
        this.artista = artista;
    }

    public void setFaixas(Integer faixas) {
        this.faixas = faixas;
    }
}
