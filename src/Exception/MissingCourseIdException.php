<?php

namespace App\Exception;

class MissingCourseIdException extends ApiException
{
    public function __construct(?\Throwable $cause = null)
    {
        parent::__construct('MISSING_COURSE_ID', 400, 'An LMS course ID was required but was not present in the request context.', [], $cause);
    }
}