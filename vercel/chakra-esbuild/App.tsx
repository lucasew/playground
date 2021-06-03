import * as React from 'react'

import { Badge, Box, ChakraProvider, Flex, SimpleGrid, Spinner, Text } from '@chakra-ui/react'

export default function RootComponent(props) {
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
