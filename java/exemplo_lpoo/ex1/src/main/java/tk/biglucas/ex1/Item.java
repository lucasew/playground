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
public abstract class Item {
    private String titulo;
    private Double duracao;
    private String comentarios;
    
    protected void middle_print() {
        System.out.printf("titulo=%s ", this.titulo);
        System.out.printf("durção=%f ", this.duracao);
        System.out.printf("comentarios=%s ", this.comentarios);
    }
    
    public void print() {
        System.out.print("Item (");
        this.middle_print();
        System.out.println(")");
    }

    public Item(String titulo, Double duracao, String comentarios) {
        this.titulo = titulo;
        this.duracao = duracao;
        this.comentarios = comentarios;
    }

    public Item() {
    }

    public String getComentarios() {
        return comentarios;
    }

    public Double getDuracao() {
        return duracao;
    }

    public String getTitulo() {
        return titulo;
    }

    public void setComentarios(String comentarios) {
        this.comentarios = comentarios;
    }

    public void setDuracao(Double duracao) {
        this.duracao = duracao;
    }

    public void setTitulo(String titulo) {
        this.titulo = titulo;
    }
}
