import React, {useState} from 'react';
import { View, Text, StyleSheet, TextInput, TouchableOpacity } from 'react-native';
import {heightPercentageToDP as hp, widthPercentageToDP as wp} from '../utils/responsive'
import logar from '../services/userService';

export default function Login() {
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const handleLogin = () => {
        const obj = {
            "email": email,
            "password": password,
        }
        logar(obj)
    }
  return (
    <View style={styles.page}>
        <View style={styles.title}>
            <Text style={styles.textTitle}>Lanchonete</Text>
            <Text style={styles.textTitle}>dos</Text>
            <Text style={styles.textTitle}>ZAMIGO</Text>
        </View>
        <View style={styles.body}>
            <View>
                <Text>Usuário</Text>
                <TextInput style={styles.inputs} onChangeText={setEmail}></TextInput>
            </View>
            <View>
                <Text>Senha</Text>
                <TextInput style={styles.inputs} onKeyPress={setPassword}></TextInput>
            </View>
            <View>
                <TouchableOpacity style={styles.entrar} onPress={handleLogin}>
                    <Text>Entrar</Text>
                </TouchableOpacity>
            </View>
        </View>
    </View>
  );
}


const styles = StyleSheet.create({
    page:{
      flex:1,
      backgroundColor:'yellow',
    },
    title:{
      alignItems:'center',
      justifyContent:"center",
      flex:2
    },
    textTitle:{
      fontSize:45
    },
    body:{
      flex:5,
      justifyContent:"center",
      alignItems:'center'
    },
    containerInputs:{
      justifyContent:"center",
      marginVertical:12
    },
    inputs:{
      backgroundColor:'white',
      width:wp("90%"),
      borderRadius:10
    },
    entrar:{
      borderColor:"grey",
      borderWidth:.5,
      borderRadius:9,
      paddingHorizontal:25,
      paddingVertical:10,
      backgroundColor:"#d1fffe",
      width:wp("90%"),
      alignItems:'center'
    },
    textTitleZ:{
      fontSize:45,
      textShadowColor:'green',
      textShadowRadius:5
    },
  })