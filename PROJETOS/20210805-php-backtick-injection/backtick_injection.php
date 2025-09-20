<?php
var_dump($argv);
$cmd = escapeshellcmd($argv[1]);
var_dump(`echo $cmd`);
?>
