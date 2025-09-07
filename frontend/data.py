from enum import IntEnum

class Kind(IntEnum):
    AI = 0
    User = 1

    def __str__(self) -> str:
        if self.value == 0:
            return "ai"
        else:
            return "user"

class Message:
    def __init__(self, kind: Kind, message: str) -> None:
        self._kind = kind
        self._message = message


    def get_kind(self) -> Kind:
        return self._kind
    
    def get_message(self) -> str:
        return self._message

class Prompt:
    def __init__(self, kind: int) -> None:
        assert kind < 2, "Not a valid option"
        self._kind = kind

    def get_kind(self) -> int:
        return self._kind
    
    def set_kind(self, index: int):
        self._kind = index
class User:
    """
    Defines the entire state of a User. A User needs to have
    """
    def __init__(self, username: str, jwt: str, is_premium: bool, is_admin: bool, *, local_only: bool = True) -> None:
        self._messages: list[Message] = []
        self._username: str = username
        self._jwt: str = jwt
        self._is_premium: bool = is_premium
        self._is_admin: bool = is_admin
        self._local_only: bool = local_only
        self._prompt: Prompt = Prompt(0)
        self._legal_libary: bool = False
        self._deep_think: bool = False

    def get_username(self) -> str:
        return self._username
    
    def get_jwt(self) -> str:
        return self._jwt
    
    def get_prompt(self) -> Prompt:
        return self._prompt
    
    def get_legal_libary(self) -> bool:
        return self._legal_libary
    
    def set_legal_libary(self, value: bool):
        self._legal_libary = value

    def is_premium(self) -> bool:
        return self._is_premium
    
    def is_admin(self) -> bool:
        return self._is_admin
    
    def set_prompt(self, new: Prompt):
        self._prompt = new
    
    def set_local(self, value: bool):
        self._local_only = value

    def get_local(self) -> bool:
        return self._local_only
    
    def get_messages(self) -> list[Message]:
        return self._messages
    
    def add_msg(self, msg: Message):
        self._messages.append(msg)

    def reset_history(self):
        self._messages = []

    def is_deep_think(self) -> bool:
        return self._deep_think
    
    def set_deep_think(self, value: bool):
        self._deep_think = value
