import React, {useState} from 'react';
import {View, Text, StyleSheet, TouchableOpacity} from 'react-native';

// import { Container } from './styles';



export default function ItemProduto(props) {
   
    const [qtd, setQtd] = useState(0) 

    const handleIncrease = () => {
        props.increase(props.product)
        setQtd(qtd+1)
    }
    
    const handleDecrease = () => {
        if(qtd > 0){
            props.decrease(props.product)
            setQtd(qtd-1)
        }             
    }

    return (
        <View style={styles.component}>
            <View>
                <Text style={styles.deionText}>{props.product.name} - {props.product.price} R$</Text>
            </View>
            <View style={styles.containerButtons}>
                
                <TouchableOpacity style={styles.button} onPress={handleDecrease}>
                    <Text style={styles.buttonText}>-</Text>
                </TouchableOpacity>

                <View>
                    <Text style={styles.deionText}>{qtd}</Text>
                </View>
                
                <TouchableOpacity style={styles.button} onPress={handleIncrease}>
                    <Text style={styles.buttonText}>+</Text>
                </TouchableOpacity>
                
            </View>
        </View>
    );
}

const styles = StyleSheet.create({
    component:{
        flex:1,
        flexDirection:'row',
        justifyContent:"space-between",
        alignItems:'center',
        marginHorizontal:8,
        borderBottomColor: 'grey',
        borderBottomWidth: .5,
    },
    containerButtons:{
        flexDirection:'row',
        alignItems:'center'
    },  
    button:{
        padding:12
    },
    buttonText:{
        color:"green",
        fontSize:45,
    },
    deionText:{
        fontSize:15
    }

})