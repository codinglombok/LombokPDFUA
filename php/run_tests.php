<?php
require __DIR__ . '/../../../phpunit_shim.php';
require __DIR__ . '/LombokPDFUA.php';
require __DIR__ . '/LombokPDFUATest.php';

[$passed, $failed, $failures] = run_shim_test_class('LombokPDFUATest');

echo "\n---\n";
echo "Passed: $passed, Failed: $failed\n";
if ($failed > 0) {
    exit(1);
}
