<?php

require_once './h5keeper.php';

$h5 = new H5keeper("h5keeper://127.0.0.1:9530");

// Set
try {
    $ret = $h5->Set("/ls/app/key", rand());
    print_r($ret);
} catch (Exception $e) {
    echo $e->getMessage();
}

// Get
try {
    $ret = $h5->Get("/ls/app/key");
    print_r($ret);
} catch (Exception $e) {
    echo $e->getMessage();
}


