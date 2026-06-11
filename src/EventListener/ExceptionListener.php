<?php

namespace App\EventListener;

use App\Exception\ApiException;
use Symfony\Component\DependencyInjection\Attribute\Autowire;
use Symfony\Component\EventDispatcher\Attribute\AsEventListener;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpKernel\Event\ExceptionEvent;
use Symfony\Component\HttpKernel\KernelEvents;

#[AsEventListener(event: KernelEvents::EXCEPTION, method: 'onKernelException', priority: 10)]
class ExceptionListener
{
    public function __construct(
        #[Autowire('%kernel.debug%')] private bool $isDebug
    ) {}

    public function onKernelException(ExceptionEvent $event): void
    {
        $exception = $event->getThrowable();

        if ($exception instanceof ApiException) {
            $errorDetails = $exception->getDetails();
            $errorCode = $exception->getErrorCode();
            $httpStatus = $exception->getCode();

            $response = [
                'errorCode' => $errorCode,
                'details' => $errorDetails,
            ];

            $event->setResponse(new JsonResponse($response, $httpStatus));
            return;
        }

        // Preserve default Symfony debug responses when error is unexpected
        if ($this->isDebug) {
            return;
        }

        $response = [
            'errorCode' => 'INTERNAL_SERVER_ERROR',
        ];
        $httpStatus = $exception->getCode();

        $event->setResponse(new JsonResponse($response, $httpStatus));
    }
}