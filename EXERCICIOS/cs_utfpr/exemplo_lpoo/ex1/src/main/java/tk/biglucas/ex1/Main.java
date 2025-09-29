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
public class Main {
    public static void main(String[] args) {
        Database db = new Database();
        db.add(new CD("Banda Djavu", 12, "Só as braba", 120., "É show"));
        db.add(new DVD("tarantino", "django livre", 200.0, "Recomendo"));
        // db.add(new Item()); // ele não deixa criar
        db.list();
        db.listCD();
        db.listDVD();
    }
    
}
