<?php

abstract class FileAnalyzerRecursive
{
    /**
     * @var RecursiveIteratorIterator
     */
    private $iterator;

    /**
     * FileAnalyzerRecursive constructor.
     * @param string $path
     */
    public function __construct(string $path)
    {
        $dirIterator = new RecursiveDirectoryIterator($path, FilesystemIterator::KEY_AS_FILENAME | FilesystemIterator::SKIP_DOTS);
        $this->iterator = new RecursiveIteratorIterator($dirIterator, RecursiveIteratorIterator::SELF_FIRST);
    }

    /**
     * Process recursive dir analyze
     */
    public function process()
    {
        foreach ($this->iterator as $file)
        {
            /** @var SplFileInfo $file */
            if ($file->isReadable() && $file->getExtension() == 'php') {
                $this->analyzeFile($file);
            }
        }
    }

    /**
     * @param SplFileInfo $file
     */
    protected function analyzeFile(SplFileInfo $file)
    {
        $className = $file->getBasename('.php');
        if (!class_exists($className, false)) {
            require_once $file->getRealPath();
        }

        if (class_exists($className, false)) {
            $this->processClass(new ReflectionClass($className));
            echo '.';
        }
    }

    /**
     * @param ReflectionClass $class
     * @return mixed
     */
    abstract protected function processClass(ReflectionClass $class);
}
