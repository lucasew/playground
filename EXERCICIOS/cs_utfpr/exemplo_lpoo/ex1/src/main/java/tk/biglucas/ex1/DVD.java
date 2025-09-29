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
public class DVD extends Item {
    private String diretor;

    public DVD(String diretor, String titulo, Double duracao, String comentarios) {
        super(titulo, duracao, comentarios);
        this.diretor = diretor;
    }

    public DVD() {
    }

    public String getDiretor() {
        return diretor;
    }

    public void setDiretor(String diretor) {
        this.diretor = diretor;
    }
    
    protected void middle_print() {
        System.out.printf("diretor=%s ", diretor);
    }
    
    public void print() {
        System.out.print("DVD(");
        super.middle_print();
        middle_print();
        System.out.println(")");
    }
}
