import * as React from 'react'
import * as Server from 'react-dom/server'

import { Badge, Box, ChakraProvider, Flex, SimpleGrid, Spinner, Text } from '@chakra-ui/react'

const port = 8080;

import Koa from 'koa'

function RootComponent(props) {
    const [loading, setLoading] = React.useState(true)
    React.useEffect(() => {
        return () => clearTimeout(setTimeout(() => setLoading(false), 3000))
    }, [])
    console.log("processando")
    if (loading) {
        return (
            <ChakraProvider>
                <Flex alignItems="center" justifyContent="center">
                    <Spinner size="xl"/>
                </Flex>
            </ChakraProvider>
        )
    }
    return (
        <ChakraProvider>
            <SimpleGrid columns={{sm: 2, md: 4}}>
                <Text fontSize="md">Hello, world</Text>
                <Text fontSize="md">Hello, world</Text>
                <Text fontSize="md">Hello, world</Text>
                <Text fontSize="md">Hello, world</Text>
                <Text fontSize="md">Hello, world</Text>
                <Text fontSize="md">Hello, world</Text>
                <Text fontSize="md">Hello, world</Text>
                <Badge colorScheme="green">Bão</Badge>
            </SimpleGrid>
        </ChakraProvider>
    )
}


const app = new Koa()
app.use(async ctx => {
    // This is vanilla SSR xD
    // BTW that print does not run on client
    ctx.body = Server.renderToString(<RootComponent/>)
})

app.listen(port)
console.log(`listening on port ${port}`)
