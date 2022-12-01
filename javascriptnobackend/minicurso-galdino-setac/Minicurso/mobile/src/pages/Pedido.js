import React from 'react';
import { View, StyleSheet } from 'react-native';

import {heightPercentageToDP as hp, widthPercentageToDP as wp} from '../utils/responsive';
import ItemProduto from '../components/ItemPedido';

export default function Pedido() {
  const [table, setTable] = useState('');
  const [produtosComanda, setProdutosComanda] = useState({})
  const 
    {
      name: "",
      price: "",
    }
  ]
  return (
    <View style={styles.page}>
      <View style={styles.containerTitlePage}>
        <Text style={styles.titlePage}>ZAMIGO</Text>
        <TextInput style={styles.inputMesa}></TextInput>
      </View>
      <View>

      </View>
      <View>
        <TouchableOpacity>
          <Text>Trazer Pedido</Text>
        </TouchableOpacity>
      </View>
      <View>
        <ScrollView>
          produtos.map(produto => <ItemPedido key={produto.name} product={produto}></ItemPedido>)
        </ScrollView>
      </View>
    </View>
  )
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