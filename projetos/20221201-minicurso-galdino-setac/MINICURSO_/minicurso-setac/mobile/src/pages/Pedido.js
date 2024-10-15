import React, {useState, useEffect} from 'react';
import { View, StyleSheet, Text, ScrollView, TextInput, TouchableOpacity } from 'react-native';

import {
    widthPercentageToDP as wp,
    heightPercentageToDP as hp
  } from "../utils/responsive";
import ItemProduto from '../components/ItemProduto';
import { buscarProdutos, salvarComanda } from '../services/produtoService';
  

// import { Container } from './styles';

export default function Pedido() {

    const [table, setTable] = useState("")
    const [produtosComanda, setProdutosComanda] = useState([])
    const [produtos, setProdutos] = useState([])

    useEffect(()=>{
        fetchProdutos()
    },[])

    const fetchProdutos = () =>{
        buscarProdutos()
        .then(response => setProdutos(response.data))
        .catch(err => console.log(err))
    }

    const increase = (product)=>{
        produtosComanda.push(product)   
        setProdutosComanda(produtosComanda)
        return product;
    }

    const decrease = (product)=>{
        const index = produtosComanda.indexOf(product);
        delete produtosComanda[index]
        setProdutosComanda(produtosComanda)            
        return product;
    }

    const handleMakeCommand = () => {

        console.log(produtosComanda);
        

        if(!table || !table.trim)
            alert('Comanda inválida!\nInforme uma mesa!')
        
        if(!produtosComanda || !produtosComanda.length)
            alert('Comanda inválida!\nInforme um produto!')

        return{
            table: table,
            products: produtosComanda
        }
    }

    const handleSaveCommand = () => {
        salvarComanda(handleMakeCommand())
            .then( _ => {
                alert('Comanda salva com sucesso!')
                setProdutosComanda([])
                setProdutos([])
                setTable("")
                fetchProdutos()
            })
            .catch(err => alert(err))
    }

    return (
        <View style={styles.page}>

            <View style={styles.containerTitlePage}>
                <Text style={styles.titlePage}>ZAMIGOS</Text>
            </View>

            <View style={styles.containerMesa}>
                <Text style={styles.textMesa}>Mesa</Text>
                <TextInput style={styles.inputMesa} onChangeText={setTable} value={table} keyboardType={"number-pad"}/>
            </View>

            <View style={styles.body}>
                <ScrollView>
                    {
                        produtos.map(produto => <ItemProduto key={produto._id} product={produto} increase={increase} decrease={decrease}/>)
                    }
                </ScrollView>
            </View>

            <View style={styles.footer}>
                <TouchableOpacity style={styles.footerButton} onPress={handleSaveCommand}>
                    <Text>Fazer Pedido</Text>
                </TouchableOpacity>
            </View>
        </View>
    );
}


const styles = StyleSheet.create({
    page:{
        flex:1,
    },
    textMesa:{
        fontSize:35
    },
    inputMesa:{
        borderColor:'grey',
        borderWidth:.4,
        borderRadius:9,
        width:wp("35")
    },
    containerMesa:{
        flexDirection:'row',
        justifyContent:"space-around",
        flex:.8,
        alignItems:"center",
    },
    containerTitlePage:{
        flex:1,
        alignItems:"center",
        justifyContent:'center'
    },
    titlePage:{
        fontSize:35,
        textShadowColor:'green',
        textShadowRadius:5,

    },
    body:{
        flex:4,
    },
    footer:{
        flex:.5,
    },
    footerButton:{
        backgroundColor:'green',
        flex:1,
        justifyContent:'center',
        alignItems:"center"

    },  
    footerButtonText:{
        fontSize:45
    }

})