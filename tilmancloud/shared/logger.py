import logging

DEFAULT_FORMATTER = logging.Formatter('%(asctime)s %(name)-12s: %(levelname)-8s %(message)s')


def get_logger(module: str) -> logging.Logger:
    _logger = logging.getLogger(module)
    _logger.setLevel(logging.DEBUG)

    console_handler = logging.StreamHandler()
    console_handler.setLevel(logging.INFO)
    console_handler.setFormatter(DEFAULT_FORMATTER)
    _logger.addHandler(console_handler)

    return _logger
