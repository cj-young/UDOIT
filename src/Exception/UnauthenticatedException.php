<?php

namespace App\Exception;

class UnauthenticatedException extends ApiException
{
    public function __construct(?\Throwable $cause = null)
    {
        parent::__construct('UNAUTHENTICATED', 401, 'Authentication requried but no authenticated user was found in the security context.', [], $cause);
    }
}