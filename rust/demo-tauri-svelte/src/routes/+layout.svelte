<script lang="ts" type="module">
  import { Container, NavbarBrand, Icon, NavItem, NavLink, Nav, Navbar } from 'sveltestrap';
  import 'bootstrap/dist/css/bootstrap.min.css';
  import 'bootstrap-icons/font/bootstrap-icons.css';
</script>

<Container>
  <Navbar>
  <Nav style="z-index: 2">
    <NavbarBrand><Icon name="android"/></NavbarBrand>
    <NavItem><NavLink href="https://google.com">Google</NavLink></NavItem>
    <NavItem><NavLink href="/teste">Teste</NavLink></NavItem>

  </Nav>
  </Navbar>
  <Container>
    <slot/>
  </Container>
</Container>
