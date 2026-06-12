<?php

namespace App\Exception;

class ForbiddenException extends ApiException
{
    public function __construct(?\Throwable $cause = null)
    {
        parent::__construct('FORBIDDEN', 403, 'User attempted to access or modify a resource they do not have sufficient permission to.', [], $cause);
    }
}