<?php

require_once '../FileAnalyzerRecursive.php';
require_once '.settings.php';

if (isset($settingsBootstrap)) {
    require_once $settingsBootstrap;
}

class ClassConstantsFinder extends FileAnalyzerRecursive
{
    /**
     * Array of classes constants
     *
     * @var array
     */
    private $result = [];

    /**
     * @return array
     */
    public function getResult()
    {
        return $this->result;
    }

    /**
     * @param array $constants
     * @return array
     */
    private function filterConstants(array $constants)
    {
        return array_filter($constants, [self::class, 'isValidConstant'], ARRAY_FILTER_USE_BOTH);
    }

    /**
     * @param $value
     * @param string $name
     * @return bool
     */
    private static function isValidConstant($value, string $name)
    {
        return (
                stripos($name, 'sess') !== false
                || stripos($name, 'prefix') !== false
                || stripos($name, 'postfix') !== false
            )
            && stripos($name, 'time') === false
            && is_scalar($value);
    }

    /**
     * @inheritdoc
     */
    protected function processClass(ReflectionClass $class)
    {
        $this->result += $this->filterConstants($class->getConstants());
    }
}


$finder = new ClassConstantsFinder($settingsPath);
$finder->process();

// Print results
var_dump($finder->getResult());